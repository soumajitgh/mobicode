package user

import (
	"context"
	"errors"

	graphmodel "github.com/soumajitgh/mobicode/graphql/model"
)

// Resolver adapts the user service for GraphQL.
type Resolver struct {
	service *Service
}

func NewResolver(service *Service) *Resolver {
	return &Resolver{service: service}
}

func (r *Resolver) User(ctx context.Context, id string) (*graphmodel.User, error) {
	user, err := r.service.GetUser(ctx, id)
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalid) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ToGraphQL(user), nil
}
