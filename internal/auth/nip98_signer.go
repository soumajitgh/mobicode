package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// CreateNIP98Authorization creates an Authorization header value for a raw HTTP request body.
// It is intentionally limited to development tooling; mobile signing stays local to the device.
func CreateNIP98Authorization(nsec, requestURL, method, body string, now time.Time) (string, error) {
	privateKey, err := PrivateKeyFromNsec(nsec)
	if err != nil {
		return "", err
	}
	payload := sha256.Sum256([]byte(body))
	event := NostrEvent{
		PubKey:    hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())),
		CreatedAt: now.Unix(),
		Kind:      nip98Kind,
		Tags:      [][]string{{"u", requestURL}, {"method", strings.ToUpper(method)}, {"payload", hex.EncodeToString(payload[:])}},
		Content:   "",
	}
	canonical, err := canonicalEvent(event)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(canonical)
	event.ID = hex.EncodeToString(hash[:])
	signature, err := schnorr.Sign(privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign NIP-98 event: %w", err)
	}
	event.Sig = hex.EncodeToString(signature.Serialize())
	raw, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("marshal NIP-98 event: %w", err)
	}
	return "Nostr " + base64.StdEncoding.EncodeToString(raw), nil
}
