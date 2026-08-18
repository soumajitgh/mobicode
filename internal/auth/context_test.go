package auth

import (
	"context"
	"testing"
)

func TestContextWithAndWithoutPrincipal(t *testing.T) {
	ctx := context.Background()
	if _, ok := PrincipalFromContext(ctx); ok {
		t.Fatalf("expected no principal in empty context")
	}

	principal := Principal{
		UserID:    "user-123",
		SessionID: "session-456",
	}

	ctxWithP := WithPrincipal(ctx, principal)
	got, ok := PrincipalFromContext(ctxWithP)
	if !ok {
		t.Fatalf("expected principal in context")
	}
	if got.UserID != principal.UserID || got.SessionID != principal.SessionID {
		t.Fatalf("got principal %+v, want %+v", got, principal)
	}
}
