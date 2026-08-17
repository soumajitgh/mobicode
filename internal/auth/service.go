package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/user"
)

const refreshTokenLifetime = 30 * 24 * time.Hour

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidRegistration = errors.New("invalid registration")
)

// Tokens are returned to a mobile client after successful authentication.
type Tokens struct {
	AccessToken  string
	RefreshToken string
	User         *user.User
}

// Service owns account registration and token issuance.
type Service struct {
	users    user.Repository
	refresh  Repository
	password *PasswordService
	jwt      *JWTService
	log      *zap.Logger
}

func NewService(users user.Repository, refresh Repository, password *PasswordService, jwt *JWTService, log *zap.Logger) *Service {
	return &Service{users: users, refresh: refresh, password: password, jwt: jwt, log: log}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (*Tokens, error) {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)
	if name == "" || !validEmail(email) || len(password) < 8 {
		return nil, ErrInvalidRegistration
	}
	hash, err := s.password.Hash(password)
	if err != nil {
		return nil, err
	}
	account := &user.User{ID: uuid.NewString(), Name: name, Email: email, PasswordHash: hash}
	if err := s.users.Create(ctx, account); err != nil {
		s.log.Error("register user", zap.Error(err))
		return nil, err
	}
	return s.issueTokens(ctx, account)
}

func (s *Service) Login(ctx context.Context, email, password string) (*Tokens, error) {
	account, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if err != nil || account.PasswordHash == "" || !s.password.Verify(password, account.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, account)
}

func (s *Service) Refresh(ctx context.Context, token string) (*Tokens, error) {
	hash := hashToken(token)
	current, err := s.refresh.FindActiveByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	account, err := s.users.FindByID(ctx, current.UserID)
	if err != nil {
		return nil, err
	}
	replacement, err := newRefreshToken(account.ID)
	if err != nil {
		return nil, err
	}
	if err := s.refresh.Rotate(ctx, hash, replacement.record); err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.Sign(account.ID)
	if err != nil {
		return nil, err
	}
	return &Tokens{AccessToken: accessToken, RefreshToken: replacement.raw, User: account}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.refresh.RevokeByHash(ctx, hashToken(token))
}

func (s *Service) issueTokens(ctx context.Context, account *user.User) (*Tokens, error) {
	refresh, err := newRefreshToken(account.ID)
	if err != nil {
		return nil, err
	}
	if err := s.refresh.Create(ctx, refresh.record); err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.Sign(account.ID)
	if err != nil {
		return nil, err
	}
	return &Tokens{AccessToken: accessToken, RefreshToken: refresh.raw, User: account}, nil
}

type issuedRefreshToken struct {
	raw    string
	record *RefreshToken
}

func newRefreshToken(userID string) (*issuedRefreshToken, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	raw := base64.RawURLEncoding.EncodeToString(bytes)
	return &issuedRefreshToken{raw: raw, record: &RefreshToken{ID: uuid.NewString(), UserID: userID, TokenHash: hashToken(raw), ExpiresAt: time.Now().Add(refreshTokenLifetime)}}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func validEmail(email string) bool { _, err := mail.ParseAddress(email); return err == nil }
