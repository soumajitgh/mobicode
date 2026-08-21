package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestPairingFlow_ConfiguresOwner(t *testing.T) {
	key := testutil.NewKeyPair(t)
	app := testutil.NewApp(t)
	started, err := app.Setup.Start(context.Background(), "")
	require.NoError(t, err)

	page := httptest.NewRequest(http.MethodGet, "/setup/", nil)
	page.AddCookie(&http.Cookie{Name: "mobicode_setup", Value: started.BrowserToken})
	pageResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(pageResponse, page)
	require.Equal(t, http.StatusOK, pageResponse.Code)
	require.Contains(t, pageResponse.Body.String(), "Pair your MobiCode mobile")
	// The page refreshes waiting-session pairing tokens. Retrieve that test-only
	// session state so the mobile request uses the same token shown in the QR code.
	started, err = app.Setup.Start(context.Background(), started.BrowserToken)
	require.NoError(t, err)

	pairBody, err := json.Marshal(map[string]string{"pairingToken": started.PairingToken})
	require.NoError(t, err)
	pair := httptest.NewRequest(http.MethodPost, "/setup/pair", bytes.NewReader(pairBody))
	pair.Header.Set("Content-Type", "application/json")
	pair.Header.Set("Authorization", testutil.SignRequest(t, key, http.MethodPost, app.Config.PublicBaseURL+"/setup/pair", pairBody))
	pairResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(pairResponse, pair)
	require.Equal(t, http.StatusAccepted, pairResponse.Code, pairResponse.Body.String())

	confirm := httptest.NewRequest(http.MethodPost, "/setup/confirm", bytes.NewBufferString(url.Values{"csrf": {started.CSRFToken}}.Encode()))
	confirm.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	confirm.AddCookie(&http.Cookie{Name: "mobicode_setup", Value: started.BrowserToken})
	confirmResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(confirmResponse, confirm)
	require.Equal(t, http.StatusOK, confirmResponse.Code, confirmResponse.Body.String())
	require.Contains(t, confirmResponse.Body.String(), "Server setup is complete")

	owner, err := app.Owner.Owner(context.Background())
	require.NoError(t, err)
	require.Equal(t, key.PublicKey, owner.PublicKey)
}
