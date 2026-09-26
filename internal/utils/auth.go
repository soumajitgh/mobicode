package utils

import (
	"crypto/subtle"
	"net/mail"
	"strings"
)

func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func ValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email && len(email) <= 254
}

func ValidPassword(password string) bool { return len(password) >= 12 && len(password) <= 128 }

func ValidRecoveryToken(token, expected string) bool {
	return expected != "" && subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
