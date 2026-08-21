package setup_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestSetupHandler_LoadsInitialPage(t *testing.T) {
	app := testutil.NewApp(t)
	res := testutil.Request(t, app.Handler, http.MethodGet, "/setup/", nil)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Pair your MobiCode mobile")
}

func TestSetupHandler_RejectsInvalidPairingRequest(t *testing.T) {
	key := testutil.NewKeyPair(t)
	app := testutil.NewApp(t)
	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/setup/pair", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", testutil.SignRequest(t, key, http.MethodPost, app.Config.PublicBaseURL+"/setup/pair", body))
	res := httptest.NewRecorder()
	app.Handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestSetupHandler_ConfiguredInstanceShowsCompletion(t *testing.T) {
	app := testutil.NewApp(t, testutil.WithOwner(testutil.NewKeyPair(t)))
	res := testutil.Request(t, app.Handler, http.MethodGet, "/setup/", nil)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Server setup is complete")
}
