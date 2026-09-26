package handlers

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	zxcvbn "github.com/boomhut/zxcvbn-go"
	"github.com/go-playground/validator/v10"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/utils"
	"github.com/soumajitgh/mobicode/internal/web/pages"
)

var formValidate = validator.New()

type initialUserForm struct {
	Email           string `validate:"required,email,max=254"`
	Password        string `validate:"required,min=12,max=128"`
	ConfirmPassword string `validate:"required,eqfield=Password"`
}

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

	rawEmail := r.PostFormValue("email")
	password := r.PostFormValue("password")
	confirmPassword := r.PostFormValue("confirm_password")

	form := initialUserForm{
		Email:           utils.NormalizeEmail(rawEmail),
		Password:        password,
		ConfirmPassword: confirmPassword,
	}

	if err := formValidate.Struct(&form); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, fe := range valErrs {
				if fe.Field() == "ConfirmPassword" {
					o.renderUserForm(w, r, rawEmail, "Passwords do not match")
					return
				}
			}
			for _, fe := range valErrs {
				switch fe.Field() {
				case "Email":
					o.renderUserForm(w, r, rawEmail, "Please enter a valid email address")
					return
				case "Password":
					o.renderUserForm(w, r, rawEmail, "Password must be at least 12 characters long")
					return
				}
			}
		}
		o.renderUserForm(w, r, rawEmail, "Invalid input")
		return
	}

	user, err := o.Auth.Service.CreateInitialUser(r.Context(), form.Email, form.Password)
	if err != nil {
		message := "Invalid input"
		if errors.Is(err, auth.ErrDuplicateEmail) {
			message = "Email already registered"
		} else if errors.Is(err, auth.ErrInitialUserAlreadyExists) {
			o.Sessions.Put(r.Context(), "onboarding_in_progress", true)
			http.Redirect(w, r, "/onboarding?step=2", http.StatusSeeOther)
			return
		}

		o.renderUserForm(w, r, rawEmail, message)
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

func (o *Onboarding) renderUserForm(w http.ResponseWriter, r *http.Request, email, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	templ.Handler(pages.Onboarding(pages.OnboardingProps{
		Step:         1,
		CSRFToken:    o.Auth.csrf(r.Context()),
		ErrorMessage: message,
		EmailValue:   email,
	})).ServeHTTP(w, r)
}

func (o *Onboarding) PasswordStrength(w http.ResponseWriter, r *http.Request) {
	if !o.Auth.checkCSRF(r) {
		http.Error(w, "invalid CSRF token", http.StatusForbidden)
		return
	}

	password := r.PostFormValue("password")
	if err := formValidate.Var(password, "max=128"); err != nil {
		http.Error(w, "password too long", http.StatusBadRequest)
		return
	}

	score := 0
	if password != "" {
		email := utils.NormalizeEmail(r.PostFormValue("email"))
		score = zxcvbn.PasswordStrength(password, []string{email, "MobiCode"}).Score
	}
	w.Header().Set("Cache-Control", "no-store")
	templ.Handler(pages.PasswordStrength(password, score)).ServeHTTP(w, r)
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
