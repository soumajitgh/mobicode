package testutil

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/auth"
)

// KeyPair is a disposable Nostr identity for tests.
type KeyPair struct {
	Nsec      string
	PublicKey string
}

// NewKeyPair creates a new disposable Nostr identity.
func NewKeyPair(t *testing.T) KeyPair {
	t.Helper()
	nsec, publicKey, err := auth.GenerateNsec()
	require.NoError(t, err)
	return KeyPair{Nsec: nsec, PublicKey: publicKey}
}

// SignRequest creates a NIP-98 Authorization header for a request body.
func SignRequest(t *testing.T, key KeyPair, method, targetURL string, body []byte) string {
	t.Helper()
	header, err := auth.CreateNIP98Authorization(key.Nsec, targetURL, method, string(body), time.Now())
	require.NoError(t, err)
	return header
}

// WithAuthorization attaches a NIP-98 proof to an HTTP request.
func WithAuthorization(req *http.Request, value string) *http.Request {
	req.Header.Set("Authorization", value)
	return req
}
