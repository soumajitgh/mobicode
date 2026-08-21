package auth

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestNIP98Verifier_Verify(t *testing.T) {
	now := time.Unix(1_780_000_000, 0)
	privateKey, _ := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{7}, 32))
	body := `{"query":"query Viewer { viewer { publicKey } }"}`
	verifier := NewNIP98Verifier(&config.Config{PublicBaseURL: "https://example.com"})
	verifier.now = func() time.Time { return now }

	tests := []struct {
		name    string
		request func(t *testing.T) *http.Request
		wantErr bool
	}{
		{name: "valid request", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now, http.MethodPost, "https://example.com/graphql", body, nil)
		}},
		{name: "invalid signature", request: func(t *testing.T) *http.Request {
			r := signedRequest(t, privateKey, now, http.MethodPost, "https://example.com/graphql", body, nil)
			return withAuthorizationEvent(t, r, func(event *NostrEvent) { event.Sig = strings.Repeat("0", 128) })
		}, wantErr: true},
		{name: "wrong request URL", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now, http.MethodPost, "https://other.example/graphql", body, nil)
		}, wantErr: true},
		{name: "wrong request method", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now, http.MethodGet, "https://example.com/graphql", body, nil)
		}, wantErr: true},
		{name: "missing payload tag", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now, http.MethodPost, "https://example.com/graphql", body, [][]string{{"u", "https://example.com/graphql"}, {"method", http.MethodPost}})
		}, wantErr: true},
		{name: "expired event", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now.Add(-proofLifetime-time.Second), http.MethodPost, "https://example.com/graphql", body, nil)
		}, wantErr: true},
		{name: "future event", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now.Add(proofLifetime+time.Second), http.MethodPost, "https://example.com/graphql", body, nil)
		}, wantErr: true},
		{name: "incorrect payload hash", request: func(t *testing.T) *http.Request {
			return signedRequest(t, privateKey, now, http.MethodPost, "https://example.com/graphql", body, [][]string{{"u", "https://example.com/graphql"}, {"method", http.MethodPost}, {"payload", strings.Repeat("0", 64)}})
		}, wantErr: true},
		{name: "modified request body", request: func(t *testing.T) *http.Request {
			r := signedRequest(t, privateKey, now, http.MethodPost, "https://example.com/graphql", body, nil)
			r.Body = io.NopCloser(strings.NewReader(`{"query":"mutation { nope }"}`))
			return r
		}, wantErr: true},
		{name: "malformed authorization header", request: func(t *testing.T) *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
			r.Header.Set("Authorization", "Nostr not-base64")
			return r
		}, wantErr: true},
		{name: "invalid encoded event", request: func(t *testing.T) *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
			r.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString([]byte(`{"kind":27235}`)))
			return r
		}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := verifier.Verify(tt.request(t))
			if tt.wantErr {
				require.ErrorIs(t, err, ErrUnauthenticated)
				return
			}
			require.NoError(t, err)
			require.Equal(t, hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())), proof.PublicKey)
		})
	}
}

func TestNIP98Verifier_RestoresRequestBody(t *testing.T) {
	now := time.Unix(1_780_000_000, 0)
	privateKey, _ := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{7}, 32))
	body := `{"query":"query Viewer { viewer { publicKey } }"}`
	request := signedRequest(t, privateKey, now, http.MethodPost, "https://example.com/graphql", body, nil)
	verifier := NewNIP98Verifier(&config.Config{PublicBaseURL: "https://example.com"})
	verifier.now = func() time.Time { return now }

	_, err := verifier.Verify(request)
	require.NoError(t, err)
	restored, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Equal(t, body, string(restored))
}

func signedRequest(t *testing.T, privateKey *btcec.PrivateKey, createdAt time.Time, method, targetURL, body string, tags [][]string) *http.Request {
	t.Helper()
	if tags == nil {
		payload := sha256.Sum256([]byte(body))
		tags = [][]string{{"u", targetURL}, {"method", method}, {"payload", hex.EncodeToString(payload[:])}}
	}
	event := NostrEvent{PubKey: hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())), CreatedAt: createdAt.Unix(), Kind: nip98Kind, Tags: tags}
	canonical, err := canonicalEvent(event)
	require.NoError(t, err)
	hash := sha256.Sum256(canonical)
	event.ID = hex.EncodeToString(hash[:])
	sig, err := schnorr.Sign(privateKey, hash[:])
	require.NoError(t, err)
	event.Sig = hex.EncodeToString(sig.Serialize())

	raw, err := json.Marshal(event)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	request.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString(raw))
	return request
}

func withAuthorizationEvent(t *testing.T, request *http.Request, mutate func(*NostrEvent)) *http.Request {
	t.Helper()
	parts := strings.Fields(request.Header.Get("Authorization"))
	raw, err := base64.StdEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var event NostrEvent
	require.NoError(t, json.Unmarshal(raw, &event))
	mutate(&event)
	raw, err = json.Marshal(event)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString(raw))
	return request
}
