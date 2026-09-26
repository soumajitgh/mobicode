package utils

import (
	"strings"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  USER@EXAMPLE.COM  ", "user@example.com"},
		{"user@example.com", "user@example.com"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := NormalizeEmail(tt.input); got != tt.expected {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"owner@example.com", true},
		{"test.user+tag@domain.co.uk", true},
		{"", false},
		{"invalid", false},
		{"@example.com", false},
		{"user@", false},
		{strings.Repeat("a", 243) + "@example.com", false}, // > 254 bytes
	}
	for _, tt := range tests {
		if got := ValidEmail(tt.email); got != tt.valid {
			t.Errorf("ValidEmail(%q) = %v, want %v", tt.email, got, tt.valid)
		}
	}
}

func TestValidPassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"", false},
		{"short", false},
		{"12345678901", false},            // 11 chars
		{"123456789012", true},            // 12 chars
		{"validpassword123", true},        // 16 chars
		{strings.Repeat("a", 128), true},  // 128 chars
		{strings.Repeat("a", 129), false}, // 129 chars
	}
	for _, tt := range tests {
		if got := ValidPassword(tt.password); got != tt.valid {
			t.Errorf("ValidPassword(len=%d) = %v, want %v", len(tt.password), got, tt.valid)
		}
	}
}

func TestValidRecoveryToken(t *testing.T) {
	tests := []struct {
		token    string
		expected string
		valid    bool
	}{
		{"token123", "token123", true},
		{"token123", "wrongtoken", false},
		{"", "expected", false},
		{"token123", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		if got := ValidRecoveryToken(tt.token, tt.expected); got != tt.valid {
			t.Errorf("ValidRecoveryToken(%q, %q) = %v, want %v", tt.token, tt.expected, got, tt.valid)
		}
	}
}
