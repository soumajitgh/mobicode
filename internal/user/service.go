package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrInvalid    = errors.New("invalid user")
	ErrEmailTaken = errors.New("email already taken")
)

// Service contains user business logic.
type Service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalid
	}
	s.log.Debug("fetching user", zap.String("user_id", id))
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.log.Debug("user not found", zap.String("user_id", id))
			return nil, ErrNotFound
		}
		s.log.Error("failed to fetch user", zap.String("user_id", id), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, ErrInvalid
	}
	return s.repo.FindByEmail(ctx, email)
}

func (s *Service) CreateUser(ctx context.Context, name, email, passwordHash string) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" {
		return nil, ErrInvalid
	}
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}
	u := &User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return u, nil
}


