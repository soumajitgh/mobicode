package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/pairing"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

type devAutoPairResponse struct {
	AccessToken string `json:"accessToken"`
	User        struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// DevAutoPair is registered only by development servers with the opt-in flag.
func DevAutoPair(users repository.UserRepository, service *pairing.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Platform string `json:"platform"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&input); err != nil || (input.Platform != "IOS" && input.Platform != "ANDROID") {
			http.Error(w, "platform must be IOS or ANDROID", http.StatusBadRequest)
			return
		}
		user, err := users.FindFirst(r.Context())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "no account available", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "could not find account", http.StatusInternalServerError)
			return
		}
		session, err := service.AutoPair(r.Context(), user.ID, "MobiCode development device", input.Platform)
		if err != nil {
			http.Error(w, "could not pair device", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		response := devAutoPairResponse{AccessToken: session.AccessToken}
		response.User.ID = session.User.UserID
		response.User.Email = session.User.Email
		_ = json.NewEncoder(w).Encode(response)
	}
}
