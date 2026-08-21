package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestDevelopmentNsecBootstrapsOwnerAndSignsRequests(t *testing.T) {
	nsec, publicKey, err := auth.GenerateNsec()
	if err != nil {
		t.Fatalf("generate nsec: %v", err)
	}
	if got, err := auth.PublicKeyFromNsec(nsec); err != nil || got != publicKey {
		t.Fatalf("decode generated nsec: got %q, %v", got, err)
	}

	db := testutil.NewDB(t)
	owner := auth.NewOwnerService(auth.NewOwnerRepository(db))
	cfg := &config.Config{Env: "development", DevNsec: nsec}
	if err := auth.BootstrapDevelopmentIdentity(cfg, owner, zap.NewNop()); err != nil {
		t.Fatalf("bootstrap development identity: %v", err)
	}
	identity, err := owner.Owner(context.Background())
	if err != nil || identity.PublicKey != publicKey {
		t.Fatalf("unexpected configured owner: %#v, %v", identity, err)
	}

	body := `{"query":"query Viewer { viewer { publicKey } }"}`
	now := time.Unix(1_780_000_000, 0)
	header, err := auth.CreateNIP98Authorization(nsec, "https://example.com/graphql", http.MethodPost, body, now)
	if err != nil {
		t.Fatalf("create authorization: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	request.Header.Set("Authorization", header)
	verifier := auth.NewNIP98Verifier(&config.Config{PublicBaseURL: "https://example.com"})
	// The public verifier uses the production clock; create a fresh proof for this integration check.
	header, err = auth.CreateNIP98Authorization(nsec, "https://example.com/graphql", http.MethodPost, body, time.Now())
	if err != nil {
		t.Fatalf("create current authorization: %v", err)
	}
	request.Header.Set("Authorization", header)
	proof, err := verifier.Verify(request)
	if err != nil || proof.PublicKey != publicKey {
		t.Fatalf("verify development authorization: %#v, %v", proof, err)
	}
}
