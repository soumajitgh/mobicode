package utils

import (
	"crypto/subtle"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func ValidEmail(email string) bool {
	return validate.Var(email, "required,email,max=254") == nil
}

func ValidPassword(password string) bool {
	return validate.Var(password, "required,min=12,max=128") == nil
}

func ValidRecoveryToken(token, expected string) bool {
	return expected != "" && subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
