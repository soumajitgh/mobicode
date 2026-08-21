package setup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/soumajitgh/mobicode/internal/auth"
)

const setupLifetime = 10 * time.Minute

var ErrInProgress = errors.New("setup is already in progress")

type StartedSession struct {
	Session      *Session
	BrowserToken string
	PairingToken string
	CSRFToken    string
}

type Service struct {
	repo  *Repository
	owner *auth.OwnerService
}

func NewService(repo *Repository, owner *auth.OwnerService) *Service {
	return &Service{repo: repo, owner: owner}
}

func (s *Service) Start(ctx context.Context, browserToken string) (*StartedSession, error) {
	if _, err := s.owner.Owner(ctx); err == nil {
		return nil, auth.ErrForbidden
	} else if !errors.Is(err, auth.ErrSetupRequired) {
		return nil, err
	}
	if browserToken != "" {
		if existing, err := s.repo.ByBrowserToken(ctx, tokenHash(browserToken)); err != nil {
			return nil, err
		} else if existing != nil {
			if existing.State != stateWaiting {
				return &StartedSession{Session: existing, BrowserToken: browserToken}, nil
			}
			pairing, err := randomToken()
			if err != nil {
				return nil, err
			}
			csrf, err := randomToken()
			if err != nil {
				return nil, err
			}
			if err := s.repo.RefreshTokens(ctx, existing.ID, tokenHash(pairing), tokenHash(csrf)); err != nil {
				return nil, err
			}
			return &StartedSession{Session: existing, BrowserToken: browserToken, PairingToken: pairing, CSRFToken: csrf}, nil
		}
	}
	if active, err := s.repo.Active(ctx); err != nil {
		return nil, err
	} else if active != nil {
		return nil, ErrInProgress
	}
	browser, err := randomToken()
	if err != nil {
		return nil, err
	}
	pairing, err := randomToken()
	if err != nil {
		return nil, err
	}
	csrf, err := randomToken()
	if err != nil {
		return nil, err
	}
	id, err := randomToken()
	if err != nil {
		return nil, err
	}
	session := &Session{ID: id, BrowserTokenHash: tokenHash(browser), PairingTokenHash: tokenHash(pairing), CSRFTokenHash: tokenHash(csrf), State: stateWaiting, ExpiresAt: time.Now().Add(setupLifetime)}
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create setup session: %w", err)
	}
	return &StartedSession{Session: session, BrowserToken: browser, PairingToken: pairing, CSRFToken: csrf}, nil
}

func (s *Service) BrowserSession(ctx context.Context, browserToken string) (*Session, error) {
	if browserToken == "" {
		return nil, nil
	}
	return s.repo.ByBrowserToken(ctx, tokenHash(browserToken))
}

func (s *Service) Pair(ctx context.Context, pairingToken, publicKey string) error {
	return s.repo.Pair(ctx, tokenHash(pairingToken), publicKey)
}

func (s *Service) PairStatus(ctx context.Context, pairingToken, publicKey string) (string, error) {
	return s.repo.PairStatus(ctx, tokenHash(pairingToken), publicKey)
}

func (s *Service) Confirm(ctx context.Context, browserToken, csrfToken string) error {
	session, err := s.BrowserSession(ctx, browserToken)
	if err != nil || session == nil || tokenHash(csrfToken) != session.CSRFTokenHash || session.State != statePending || session.CandidatePublicKey == "" {
		return errors.New("setup confirmation is unavailable")
	}
	if err := s.owner.Configure(ctx, session.CandidatePublicKey); err != nil {
		return err
	}
	return s.repo.Complete(ctx, session.ID)
}

func (s *Service) Reset(ctx context.Context) error {
	if err := s.owner.Reset(ctx); err != nil {
		return err
	}
	return s.repo.DeleteAll(ctx)
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
