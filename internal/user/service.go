package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNotFound = errors.New("user not found")
	ErrInvalid  = errors.New("invalid user")
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
	s.log.Info("fetching user", zap.String("user_id", id))
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Error("failed to fetch user", zap.String("user_id", id), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, name, email string) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" {
		return nil, ErrInvalid
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalid
	}

	user := &User{ID: uuid.NewString(), Name: name, Email: email}
	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error("failed to create user", zap.Error(err))
		return nil, err
	}
	return user, nil
}
