package setup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"
	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
)

const setupCookie = "mobicode_setup"
const setupCSRFCookie = "mobicode_setup_csrf"

type Handler struct {
	service  *Service
	verifier *auth.NIP98Verifier
	cfg      *config.Config
}

func NewHandler(service *Service, verifier *auth.NIP98Verifier, cfg *config.Config) (*Handler, error) {
	return &Handler{service: service, verifier: verifier, cfg: cfg}, nil
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.page)
	r.Get("/status", h.status)
	r.Post("/confirm", h.confirm)
	r.Post("/pair", h.pair)
	r.Post("/pair/status", h.pairStatus)
	return r
}

func (h *Handler) page(w http.ResponseWriter, r *http.Request) {
	started, err := h.service.Start(r.Context(), h.browserToken(r))
	if errors.Is(err, ErrInProgress) {
		h.render(r.Context(), w, ProgressPage())
		return
	}
	if errors.Is(err, auth.ErrForbidden) {
		h.render(r.Context(), w, CompletePage())
		return
	}
	if err != nil {
		http.Error(w, "setup unavailable", http.StatusInternalServerError)
		return
	}
	if started.PairingToken == "" { // The completed browser session can only display status.
		h.render(r.Context(), w, ResumePage(h.csrfToken(r)))
		return
	}
	h.setBrowserCookie(w, started.BrowserToken)
	h.setCSRFCookie(w, started.CSRFToken)
	deepLink := "mobicode://pair?server=" + url.QueryEscape(h.cfg.PublicBaseURL) + "&token=" + started.PairingToken
	png, err := qrcode.Encode(deepLink, qrcode.Medium, 280)
	if err != nil {
		http.Error(w, "render pairing code", http.StatusInternalServerError)
		return
	}
	h.render(r.Context(), w, SetupPage(templ.SafeURL("data:image/png;base64,"+base64.StdEncoding.EncodeToString(png)), started.CSRFToken))
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.BrowserSession(r.Context(), h.browserToken(r))
	if err != nil || session == nil {
		h.render(r.Context(), w, ProgressPage())
		return
	}
	if session.State == statePending {
		h.render(r.Context(), w, ConfirmFragment(fingerprint(session.CandidatePublicKey)))
		return
	}
	h.render(r.Context(), w, WaitingFragment())
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Confirm(r.Context(), h.browserToken(r), r.FormValue("csrf")); err != nil {
		http.Error(w, "setup confirmation unavailable", http.StatusForbidden)
		return
	}
	h.clearBrowserCookies(w)
	h.render(r.Context(), w, CompletePage())
}

func (h *Handler) pair(w http.ResponseWriter, r *http.Request) {
	proof, err := h.verifier.Verify(r)
	if err != nil {
		http.Error(w, "authentication proof invalid", http.StatusUnauthorized)
		return
	}
	if err := h.service.owner.ClaimReplay(r.Context(), proof.ID, proof.ExpiresAt); err != nil {
		http.Error(w, "authentication proof already used", http.StatusUnauthorized)
		return
	}
	var input struct {
		PairingToken string `json:"pairingToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.PairingToken == "" {
		http.Error(w, "invalid pairing request", http.StatusBadRequest)
		return
	}
	if err := h.service.Pair(r.Context(), input.PairingToken, proof.PublicKey); err != nil {
		http.Error(w, "pairing unavailable", http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "pending_confirmation"})
}

func (h *Handler) pairStatus(w http.ResponseWriter, r *http.Request) {
	proof, err := h.verifier.Verify(r)
	if err != nil {
		http.Error(w, "authentication proof invalid", http.StatusUnauthorized)
		return
	}
	if err := h.service.owner.ClaimReplay(r.Context(), proof.ID, proof.ExpiresAt); err != nil {
		http.Error(w, "authentication proof already used", http.StatusUnauthorized)
		return
	}
	var input struct {
		PairingToken string `json:"pairingToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.PairingToken == "" {
		http.Error(w, "invalid pairing request", http.StatusBadRequest)
		return
	}
	state, err := h.service.PairStatus(r.Context(), input.PairingToken, proof.PublicKey)
	if err != nil {
		http.Error(w, "pairing unavailable", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": state})
}

func (h *Handler) render(ctx context.Context, w http.ResponseWriter, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(ctx, w); err != nil {
		http.Error(w, "render setup page", http.StatusInternalServerError)
	}
}
func (h *Handler) browserToken(r *http.Request) string {
	cookie, err := r.Cookie(setupCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}
func (h *Handler) csrfToken(r *http.Request) string {
	cookie, err := r.Cookie(setupCSRFCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}
func (h *Handler) setBrowserCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: setupCookie, Value: token, Path: "/setup", HttpOnly: true, Secure: strings.HasPrefix(h.cfg.PublicBaseURL, "https://"), SameSite: http.SameSiteStrictMode, MaxAge: int(setupLifetime.Seconds())})
}
func (h *Handler) setCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: setupCSRFCookie, Value: token, Path: "/setup", Secure: strings.HasPrefix(h.cfg.PublicBaseURL, "https://"), SameSite: http.SameSiteStrictMode, MaxAge: int(setupLifetime.Seconds())})
}
func (h *Handler) clearBrowserCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: setupCookie, Value: "", Path: "/setup", HttpOnly: true, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: setupCSRFCookie, Value: "", Path: "/setup", MaxAge: -1})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fingerprint(key string) string {
	if len(key) < 16 {
		return key
	}
	return key[:8] + "…" + key[len(key)-8:]
}
