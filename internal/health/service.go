package health

import "context"

// Service provides the application's health status.
type Service struct{}

func (s *Service) Status(_ context.Context) string {
	return "ok"
}
