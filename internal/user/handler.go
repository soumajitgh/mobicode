package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler serves user HTTP requests.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches user routes to the shared router.
func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Get("/users/{id}", h.getUser)
	r.Post("/users", h.createUser)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetUser(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(r.Context(), request.Name, request.Email)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
