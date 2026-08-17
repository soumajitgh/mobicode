package user

import graphmodel "github.com/soumajitgh/mobicode/graphql/model"

// ToGraphQL maps the persistence model to the GraphQL transport model.
func ToGraphQL(user *User) *graphmodel.User {
	if user == nil {
		return nil
	}
	return &graphmodel.User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
