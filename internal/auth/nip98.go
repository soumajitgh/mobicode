package auth

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/soumajitgh/mobicode/internal/config"
)

const (
	nip98Kind     = 27235
	maxAuthBody   = 1 << 20
	maxAuthHeader = 16 << 10
	proofLifetime = time.Minute
)

// NostrEvent is the NIP-01 event envelope used by NIP-98.
type NostrEvent struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

type VerifiedEvent struct {
	ID        string
	PublicKey string
	ExpiresAt time.Time
}

// NIP98Verifier validates a signed Nostr proof against the request it protects.
type NIP98Verifier struct {
	publicBaseURL string
	now           func() time.Time
}

func NewNIP98Verifier(cfg *config.Config) *NIP98Verifier {
	return &NIP98Verifier{publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"), now: time.Now}
}

func (v *NIP98Verifier) Verify(r *http.Request) (VerifiedEvent, error) {
	if len(r.Header.Get("Authorization")) > maxAuthHeader {
		return VerifiedEvent{}, ErrUnauthenticated
	}
	body, err := readAndRestoreBody(r)
	if err != nil {
		return VerifiedEvent{}, ErrUnauthenticated
	}
	event, err := parseAuthorization(r.Header.Get("Authorization"))
	if err != nil {
		return VerifiedEvent{}, ErrUnauthenticated
	}
	if err := v.verifyEvent(event, r, body); err != nil {
		return VerifiedEvent{}, ErrUnauthenticated
	}
	return VerifiedEvent{ID: event.ID, PublicKey: strings.ToLower(event.PubKey), ExpiresAt: time.Unix(event.CreatedAt, 0).Add(proofLifetime)}, nil
}

func readAndRestoreBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxAuthBody+1))
	if err != nil || len(body) > maxAuthBody {
		return nil, fmt.Errorf("read request body")
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func parseAuthorization(header string) (NostrEvent, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Nostr") {
		return NostrEvent{}, fmt.Errorf("invalid authorization header")
	}
	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(raw) == 0 || len(raw) > maxAuthHeader {
		return NostrEvent{}, fmt.Errorf("invalid authorization event")
	}
	var event NostrEvent
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return NostrEvent{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return NostrEvent{}, fmt.Errorf("authorization event has trailing data")
	}
	return event, nil
}

func (v *NIP98Verifier) verifyEvent(event NostrEvent, r *http.Request, body []byte) error {
	if event.Kind != nip98Kind || !validPublicKey(event.PubKey) || !validHex(event.ID, 32) || !validHex(event.Sig, 64) {
		return fmt.Errorf("invalid nostr event")
	}
	createdAt := time.Unix(event.CreatedAt, 0)
	if createdAt.Before(v.now().Add(-proofLifetime)) || createdAt.After(v.now().Add(proofLifetime)) {
		return fmt.Errorf("expired nostr event")
	}
	canonical, err := canonicalEvent(event)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(canonical)
	if !strings.EqualFold(event.ID, hex.EncodeToString(hash[:])) {
		return fmt.Errorf("nostr event id mismatch")
	}
	pubkey, err := schnorr.ParsePubKey(mustDecodeHex(event.PubKey))
	if err != nil {
		return err
	}
	sig, err := schnorr.ParseSignature(mustDecodeHex(event.Sig))
	if err != nil || !sig.Verify(hash[:], pubkey) {
		return fmt.Errorf("invalid nostr signature")
	}
	tags := tagValues(event.Tags)
	expectedURL := v.publicBaseURL + r.URL.RequestURI()
	if !singleTagEquals(tags, "u", expectedURL) || !singleTagEquals(tags, "method", r.Method) {
		return fmt.Errorf("nostr request binding mismatch")
	}
	payload := sha256.Sum256(body)
	if !singleTagEquals(tags, "payload", hex.EncodeToString(payload[:])) {
		return fmt.Errorf("nostr payload binding mismatch")
	}
	return nil
}

func canonicalEvent(event NostrEvent) ([]byte, error) {
	value := []any{0, strings.ToLower(event.PubKey), event.CreatedAt, event.Kind, event.Tags, event.Content}
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil
}

func tagValues(tags [][]string) map[string][]string {
	result := make(map[string][]string)
	for _, tag := range tags {
		if len(tag) == 2 {
			result[tag[0]] = append(result[tag[0]], tag[1])
		}
	}
	return result
}

func singleTagEquals(tags map[string][]string, key, expected string) bool {
	values := tags[key]
	return len(values) == 1 && values[0] == expected
}

func validHex(value string, bytes int) bool {
	if len(value) != bytes*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validPublicKey(value string) bool  { return validHex(value, 32) }
func mustDecodeHex(value string) []byte { decoded, _ := hex.DecodeString(value); return decoded }
