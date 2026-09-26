package pairing

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

var (
	ErrAddressUnavailable = errors.New("server address unavailable; configure MOBICODE_SERVER_BASE_URL")
	ErrValidation         = errors.New("invalid device metadata")
)

type Service struct {
	Repository repository.MobileRepository
	BaseURL    string
	Now        func() time.Time
}

type Pairing struct {
	ID        string
	QRPayload string
	ExpiresAt time.Time
}

type Session struct {
	AccessToken string
	User        repository.MobileIdentity
	Device      repository.MobileDevice
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) Create(ctx context.Context, userID uint) (*Pairing, error) {
	if s.BaseURL == "" {
		return nil, ErrAddressUnavailable
	}
	id, err := randomToken(16)
	if err != nil {
		return nil, err
	}
	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	now := s.now()
	expires := now.Add(config.PairingLifetime)
	if err := s.Repository.CreatePairing(ctx, id, userID, HashToken(token), now, expires); err != nil {
		return nil, err
	}
	q := url.Values{"server": {s.BaseURL}, "token": {token}}
	return &Pairing{ID: id, QRPayload: "mobicode://pair?" + q.Encode(), ExpiresAt: expires}, nil
}

func (s *Service) Claim(ctx context.Context, token, name, platform string) (*Session, error) {
	if token == "" {
		return nil, repository.ErrInvalidPairing
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 || (platform != "IOS" && platform != "ANDROID") {
		return nil, ErrValidation
	}
	deviceID, err := randomToken(16)
	if err != nil {
		return nil, err
	}
	sessionID, err := randomToken(16)
	if err != nil {
		return nil, err
	}
	accessToken, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	identity, device, err := s.Repository.ClaimPairing(ctx, HashToken(token), deviceID, name, platform, sessionID, HashToken(accessToken), s.now())
	if err != nil {
		return nil, err
	}
	return &Session{AccessToken: accessToken, User: identity, Device: device}, nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
