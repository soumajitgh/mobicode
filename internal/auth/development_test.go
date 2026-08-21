package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestDevelopmentNsecBootstrapsOwnerAndSignsRequests(t *testing.T) {
	nsec, publicKey, err := GenerateNsec()
	if err != nil {
		t.Fatalf("generate nsec: %v", err)
	}
	if got, err := PublicKeyFromNsec(nsec); err != nil || got != publicKey {
		t.Fatalf("decode generated nsec: got %q, %v", got, err)
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&OwnerIdentity{}); err != nil {
		t.Fatalf("migrate owner identity: %v", err)
	}
	owner := NewOwnerService(NewOwnerRepository(db))
	cfg := &config.Config{Env: "development", DevNsec: nsec}
	if err := BootstrapDevelopmentIdentity(cfg, owner, zap.NewNop()); err != nil {
		t.Fatalf("bootstrap development identity: %v", err)
	}
	identity, err := owner.Owner(context.Background())
	if err != nil || identity.PublicKey != publicKey {
		t.Fatalf("unexpected configured owner: %#v, %v", identity, err)
	}

	body := `{"query":"query Viewer { viewer { publicKey } }"}`
	now := time.Unix(1_780_000_000, 0)
	header, err := CreateNIP98Authorization(nsec, "https://example.com/graphql", http.MethodPost, body, now)
	if err != nil {
		t.Fatalf("create authorization: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	request.Header.Set("Authorization", header)
	verifier := NewNIP98Verifier(&config.Config{PublicBaseURL: "https://example.com"})
	verifier.now = func() time.Time { return now }
	proof, err := verifier.Verify(request)
	if err != nil || proof.PublicKey != publicKey {
		t.Fatalf("verify development authorization: %#v, %v", proof, err)
	}
}
