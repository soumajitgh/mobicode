package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

var (
	ErrDuplicateEmail       = errors.New("email already registered")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidRecoveryToken = errors.New("invalid recovery token")
	ErrInvalidInput         = errors.New("invalid input")
)

type Service struct {
	Users         repository.UserRepository
	RecoveryToken string
}

func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email && len(email) <= 254
}
func validPassword(password string) bool { return len(password) >= 12 && len(password) <= 128 }
func Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func Verify(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=65536,t=3,p=4" {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	hash, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil || len(salt) != 16 || len(hash) != 32 {
		return false
	}
	calculated := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, calculated) == 1
}

func (s *Service) recoveryValid(token string) bool {
	return s.RecoveryToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.RecoveryToken)) == 1
}

func (s *Service) Register(ctx context.Context, email, password, token string) (*model.User, error) {
	if !s.recoveryValid(token) {
		return nil, ErrInvalidRecoveryToken
	}
	email = NormalizeEmail(email)
	if !validEmail(email) || !validPassword(password) {
		return nil, ErrInvalidInput
	}
	hash, err := Hash(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{Email: email, PasswordHash: hash, SessionVersion: 1}
	if err := s.Users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*model.User, error) {
	email = NormalizeEmail(email)
	if !validEmail(email) {
		return nil, ErrInvalidCredentials
	}
	user, err := s.Users.FindByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !Verify(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) ResetPassword(ctx context.Context, email, password, token string) error {
	if !s.recoveryValid(token) {
		return ErrInvalidRecoveryToken
	}
	email = NormalizeEmail(email)
	if !validEmail(email) || !validPassword(password) {
		return ErrInvalidInput
	}
	user, err := s.Users.FindByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	hash, err := Hash(password)
	if err != nil {
		return err
	}
	return s.Users.ReplacePassword(ctx, user.ID, hash)
}
