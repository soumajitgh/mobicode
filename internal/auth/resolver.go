package auth

import (
	"context"

	graphmodel "github.com/soumajitgh/mobicode/graph/model"
	"github.com/soumajitgh/mobicode/internal/user"
)

// Resolver adapts the auth service for GraphQL.
type Resolver struct{ service *Service }

func NewResolver(service *Service) *Resolver { return &Resolver{service: service} }

func (r *Resolver) Register(ctx context.Context, name, email, password string) (*graphmodel.AuthPayload, error) {
	tokens, err := r.service.Register(ctx, name, email, password)
	if err != nil {
		return nil, err
	}
	return toGraphQL(tokens), nil
}

func (r *Resolver) Login(ctx context.Context, email, password string) (*graphmodel.AuthPayload, error) {
	tokens, err := r.service.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}
	return toGraphQL(tokens), nil
}

func (r *Resolver) RefreshToken(ctx context.Context, refreshToken string) (*graphmodel.AuthPayload, error) {
	tokens, err := r.service.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	return toGraphQL(tokens), nil
}

func (r *Resolver) Logout(ctx context.Context, refreshToken string) (bool, error) {
	if err := r.service.Logout(ctx, refreshToken); err != nil {
		return false, err
	}
	return true, nil
}

func toGraphQL(tokens *Tokens) *graphmodel.AuthPayload {
	return &graphmodel.AuthPayload{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, User: user.ToGraphQL(tokens.User)}
}
