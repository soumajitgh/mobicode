package health_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestHandler_Health(t *testing.T) {
	res := testutil.Request(t, http.HandlerFunc(health.NewHandler().Check), http.MethodGet, "/health", nil)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "ok", res.Body.String())
}
