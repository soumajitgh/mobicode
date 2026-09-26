package handlers

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"

	"github.com/soumajitgh/mobicode/internal/store/repository"
)

type Onboarding struct {
	Auth     *Auth
	Users    repository.UserRepository
	Sessions *scs.SessionManager
}

func NewOnboarding(auth *Auth, users repository.UserRepository, sessions *scs.SessionManager) *Onboarding {
	return &Onboarding{
		Auth:     auth,
		Users:    users,
		Sessions: sessions,
	}
}

func (o *Onboarding) IsWizardInProgress(r *http.Request) bool {
	return o.Sessions.GetBool(r.Context(), "onboarding_in_progress")
}

func (o *Onboarding) HasUsers(ctx context.Context) (bool, error) {
	count, err := o.Users.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
