package graphql_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestGraphQL_ViewerWithAuthenticatedOwner(t *testing.T) {
	owner := testutil.NewKeyPair(t)
	app := testutil.NewApp(t, testutil.WithOwner(owner))
	query := `query Viewer { viewer { publicKey } }`
	body := testutil.GraphQLBody(t, query, nil)
	authorization := testutil.SignRequest(t, owner, "POST", app.Config.PublicBaseURL+"/graphql", body)

	response := testutil.GraphQL(t, app.Handler, query, nil, testutil.WithHeader("Authorization", authorization))
	require.Empty(t, response.Errors)

	var data struct {
		Viewer struct {
			PublicKey string `json:"publicKey"`
		} `json:"viewer"`
	}
	require.NoError(t, json.Unmarshal(response.Data, &data))
	require.Equal(t, owner.PublicKey, data.Viewer.PublicKey)
}
