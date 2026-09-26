package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"

	"github.com/soumajitgh/mobicode/internal/http/middleware"
)

func (o *Onboarding) RegisterRoutes(r chi.Router) {
	limiter := middleware.NewKeyedRateLimiter(rate.Every(6*time.Second), 10, 2*time.Minute)
	limitOnboarding := middleware.RateLimit(limiter, func(r *http.Request) string {
		return r.Method + ":" + r.URL.Path + ":" + middleware.ClientIP(r)
	})

	r.Get("/onboarding", o.Get)
	r.With(limitOnboarding).Post("/onboarding", o.SubmitUser)
	r.With(limitOnboarding).Post("/onboarding/finish", o.SubmitFinish)
}
