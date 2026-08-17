// Package health provides a liveness endpoint.
package health

import "net/http"

// Handler serves health-check requests.
type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
