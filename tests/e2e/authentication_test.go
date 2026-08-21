package e2e

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestAuthenticationFlow_AuthenticatedGraphQLRequestCannotBeReplayed(t *testing.T) {
	owner := testutil.NewKeyPair(t)
	app := testutil.NewApp(t, testutil.WithOwner(owner))
	body := testutil.GraphQLBody(t, `query Viewer { viewer { publicKey } }`, nil)
	authorization := testutil.SignRequest(t, owner, http.MethodPost, app.Config.PublicBaseURL+"/graphql", body)

	first := graphqlRequest(body, authorization)
	firstResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(firstResponse, first)
	require.Equal(t, http.StatusOK, firstResponse.Code, firstResponse.Body.String())
	require.Contains(t, firstResponse.Body.String(), owner.PublicKey)

	replay := graphqlRequest(body, authorization)
	replayResponse := httptest.NewRecorder()
	app.Handler.ServeHTTP(replayResponse, replay)
	require.Equal(t, http.StatusUnauthorized, replayResponse.Code)
	require.Contains(t, replayResponse.Body.String(), "already used")
}

func graphqlRequest(body []byte, authorization string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	return req
}
