package graphql

import (
	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/store"
)

// Resolver holds services used by GraphQL resolvers.
// Add service dependencies here as the API grows.
type Resolver struct {
	HealthService *health.Service
	Store         *store.Store
}
