package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestLoad_AppliesDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("PUBLIC_BASE_URL", "")
	t.Setenv("DEV_NSEC", "")

	got, err := Load("")
	require.NoError(t, err)
	want := &Config{
		Port:          "8080",
		Env:           "development",
		DatabasePath:  "data/app.db",
		PublicBaseURL: "http://localhost:8080",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("config mismatch (-want +got):\n%s", diff)
	}
}

func TestLoad_RejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
	}{
		{name: "public URL includes a path", values: map[string]string{"PUBLIC_BASE_URL": "https://example.com/api"}},
		{name: "production uses HTTP", values: map[string]string{"ENV": "production", "PUBLIC_BASE_URL": "http://example.com"}},
		{name: "development identity outside development", values: map[string]string{"ENV": "test", "DEV_NSEC": "nsec1example"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", "")
			t.Setenv("ENV", "")
			t.Setenv("APP_ENV", "")
			t.Setenv("DATABASE_PATH", "")
			t.Setenv("PUBLIC_BASE_URL", "")
			t.Setenv("DEV_NSEC", "")
			for key, value := range tt.values {
				t.Setenv(key, value)
			}
			_, err := Load("")
			require.Error(t, err)
		})
	}
}
