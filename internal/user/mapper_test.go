package user

import (
	"testing"
	"time"
)

func TestToGraphQL(t *testing.T) {
	createdAt := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	mapped := ToGraphQL(&User{ID: "user-1", Name: "Zoravix", Email: "z@fotopick.in", CreatedAt: createdAt})

	if mapped.ID != "user-1" || mapped.Name != "Zoravix" || mapped.Email != "z@fotopick.in" || !mapped.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected mapped user: %#v", mapped)
	}
}
