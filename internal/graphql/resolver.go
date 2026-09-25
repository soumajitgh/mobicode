package graphql

import "github.com/soumajitgh/mobicode/internal/health"

// Resolver holds services used by GraphQL resolvers.
// Add service dependencies here as the API grows.
type Resolver struct {
	HealthService *health.Service
}
