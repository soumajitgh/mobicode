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
	"github.com/soumajitgh/mobicode/internal/config"
)

func TestNIP98VerifierBindsSignatureAndBody(t *testing.T) {
	now := time.Unix(1_780_000_000, 0)
	privateKey, _ := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{7}, 32))
	body := `{"query":"query Viewer { viewer { publicKey } }"}`
	request := signedRequest(t, privateKey, now, "https://example.com/graphql", body)
	verifier := NewNIP98Verifier(&config.Config{PublicBaseURL: "https://example.com"})
	verifier.now = func() time.Time { return now }

	proof, err := verifier.Verify(request)
	if err != nil {
		t.Fatalf("verify valid request: %v", err)
	}
	if proof.PublicKey != hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())) {
		t.Fatalf("unexpected public key %q", proof.PublicKey)
	}
	restored, err := io.ReadAll(request.Body)
	if err != nil || string(restored) != body {
		t.Fatalf("request body was not restored: %q, %v", restored, err)
	}

	tampered := signedRequest(t, privateKey, now, "https://example.com/graphql", body)
	tampered.Body = io.NopCloser(strings.NewReader(`{"query":"mutation { nope }"}`))
	if _, err := verifier.Verify(tampered); err == nil {
		t.Fatal("tampered body was accepted")
	}
}

func signedRequest(t *testing.T, privateKey *btcec.PrivateKey, createdAt time.Time, url, body string) *http.Request {
	t.Helper()
	payload := sha256.Sum256([]byte(body))
	event := NostrEvent{PubKey: hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())), CreatedAt: createdAt.Unix(), Kind: nip98Kind, Tags: [][]string{{"u", url}, {"method", http.MethodPost}, {"payload", hex.EncodeToString(payload[:])}}, Content: ""}
	canonical, err := canonicalEvent(event)
	if err != nil {
		t.Fatalf("canonical event: %v", err)
	}
	hash := sha256.Sum256(canonical)
	event.ID = hex.EncodeToString(hash[:])
	sig, err := schnorr.Sign(privateKey, hash[:])
	if err != nil {
		t.Fatalf("sign event: %v", err)
	}
	event.Sig = hex.EncodeToString(sig.Serialize())
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(body))
	request.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString(raw))
	return request
}
