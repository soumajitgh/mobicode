package handlers

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/utils"
	"github.com/soumajitgh/mobicode/internal/web/pages"
)

func (o *Onboarding) Get(w http.ResponseWriter, r *http.Request) {
	hasUsers, err := o.HasUsers(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if hasUsers && !o.IsWizardInProgress(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	stepParam := r.URL.Query().Get("step")
	var step int
	userEmail := ""

	if !hasUsers {
		if stepParam != "" && stepParam != "1" {
			http.Redirect(w, r, "/onboarding?step=1", http.StatusSeeOther)
			return
		}
		step = 1
	} else {
		// Users already exist and wizard is in progress
		firstUser, _ := o.Users.FindFirst(r.Context())
		if firstUser != nil {
			userEmail = firstUser.Email
		}

		switch stepParam {
		case "1":
			step = 1
		case "2", "":
			step = 2
		case "3":
			step = 3
		default:
			http.Redirect(w, r, "/onboarding?step=2", http.StatusSeeOther)
			return
		}
	}

	props := pages.OnboardingProps{
		Step:         step,
		UserCreated:  hasUsers,
		UserEmail:    userEmail,
		CSRFToken:    o.Auth.csrf(r.Context()),
		ErrorMessage: "",
		EmailValue:   "",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templ.Handler(pages.Onboarding(props)).ServeHTTP(w, r)
}

func (o *Onboarding) SubmitUser(w http.ResponseWriter, r *http.Request) {
	if !o.Auth.checkCSRF(r) {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}

	hasUsers, err := o.HasUsers(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if hasUsers {
		o.Sessions.Put(r.Context(), "onboarding_in_progress", true)
		http.Redirect(w, r, "/onboarding?step=2", http.StatusSeeOther)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	user, err := o.Auth.Service.CreateInitialUser(r.Context(), email, password)
	if err != nil {
		message := "Invalid input"
		if !utils.ValidEmail(utils.NormalizeEmail(email)) {
			message = "Please enter a valid email address"
		} else if !utils.ValidPassword(password) {
			message = "Password must be at least 12 characters long"
		} else if errors.Is(err, auth.ErrDuplicateEmail) {
			message = "Email already registered"
		} else if errors.Is(err, auth.ErrInitialUserAlreadyExists) {
			o.Sessions.Put(r.Context(), "onboarding_in_progress", true)
			http.Redirect(w, r, "/onboarding?step=2", http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		props := pages.OnboardingProps{
			Step:         1,
			UserCreated:  false,
			UserEmail:    "",
			CSRFToken:    o.Auth.csrf(r.Context()),
			ErrorMessage: message,
			EmailValue:   email,
		}
		templ.Handler(pages.Onboarding(props)).ServeHTTP(w, r)
		return
	}

	if err := o.Sessions.RenewToken(r.Context()); err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	o.Sessions.Put(r.Context(), "user_id", int(user.ID))
	o.Sessions.Put(r.Context(), "session_version", user.SessionVersion)
	o.Sessions.Put(r.Context(), "onboarding_in_progress", true)

	http.Redirect(w, r, "/onboarding?step=2", http.StatusSeeOther)
}

func (o *Onboarding) SubmitFinish(w http.ResponseWriter, r *http.Request) {
	if !o.Auth.checkCSRF(r) {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}

	hasUsers, err := o.HasUsers(r.Context())
	if err != nil || !hasUsers {
		http.Redirect(w, r, "/onboarding?step=1", http.StatusSeeOther)
		return
	}

	if o.Sessions.GetInt(r.Context(), "user_id") == 0 {
		firstUser, err := o.Users.FindFirst(r.Context())
		if err == nil && firstUser != nil {
			o.Sessions.Put(r.Context(), "user_id", int(firstUser.ID))
			o.Sessions.Put(r.Context(), "session_version", firstUser.SessionVersion)
		}
	}

	o.Sessions.Remove(r.Context(), "onboarding_in_progress")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
