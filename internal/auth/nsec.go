package auth

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/bech32"
)

// PrivateKeyFromNsec decodes a standard NIP-19 nsec private key.
func PrivateKeyFromNsec(nsec string) (*btcec.PrivateKey, error) {
	hrp, data, err := bech32.Decode(strings.TrimSpace(nsec))
	if err != nil || !strings.EqualFold(hrp, "nsec") {
		return nil, fmt.Errorf("DEV_NSEC must be a valid nsec")
	}
	secretBytes, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil || len(secretBytes) != 32 {
		return nil, fmt.Errorf("DEV_NSEC must contain a 32-byte private key")
	}
	privateKey, _ := btcec.PrivKeyFromBytes(secretBytes)
	if !bytes.Equal(privateKey.Serialize(), secretBytes) {
		return nil, fmt.Errorf("DEV_NSEC has an invalid private key")
	}
	return privateKey, nil
}

func PublicKeyFromNsec(nsec string) (string, error) {
	privateKey, err := PrivateKeyFromNsec(nsec)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())), nil
}

// GenerateNsec creates a new NIP-19 private key for local development.
func GenerateNsec() (string, string, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("generate private key: %w", err)
	}
	nsec, err := bech32.EncodeFromBase256("nsec", privateKey.Serialize())
	if err != nil {
		return "", "", fmt.Errorf("encode nsec: %w", err)
	}
	return nsec, hex.EncodeToString(schnorr.SerializePubKey(privateKey.PubKey())), nil
}
