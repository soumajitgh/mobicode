package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestRequireOwner_RejectsNonOwner(t *testing.T) {
	owner := testutil.NewKeyPair(t)
	other := testutil.NewKeyPair(t)
	app := testutil.NewApp(t, testutil.WithOwner(owner))
	body := testutil.GraphQLBody(t, `query Viewer { viewer { publicKey } }`, nil)
	req, err := http.NewRequest(http.MethodPost, "http://mobicode.test/graphql", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	testutil.WithAuthorization(req, testutil.SignRequest(t, other, http.MethodPost, app.Config.PublicBaseURL+"/graphql", body))

	res := httptest.NewRecorder()
	app.Handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusForbidden, res.Code)
}
