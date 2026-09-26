package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
	"github.com/soumajitgh/mobicode/internal/utils"
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

func (s *Service) Register(ctx context.Context, email, password, token string) (*model.User, error) {
	if !utils.ValidRecoveryToken(token, s.RecoveryToken) {
		return nil, ErrInvalidRecoveryToken
	}
	email = utils.NormalizeEmail(email)
	if !utils.ValidEmail(email) || !utils.ValidPassword(password) {
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
	email = utils.NormalizeEmail(email)
	if !utils.ValidEmail(email) {
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
	if !utils.ValidRecoveryToken(token, s.RecoveryToken) {
		return ErrInvalidRecoveryToken
	}
	email = utils.NormalizeEmail(email)
	if !utils.ValidEmail(email) || !utils.ValidPassword(password) {
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
