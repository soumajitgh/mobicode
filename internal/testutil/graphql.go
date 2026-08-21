package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string         `json:"message"`
	Path    []any          `json:"path,omitempty"`
	Ext     map[string]any `json:"extensions,omitempty"`
}

type RequestOption func(*http.Request)

// WithHeader adds an HTTP header to a GraphQL request.
func WithHeader(name, value string) RequestOption {
	return func(req *http.Request) { req.Header.Set(name, value) }
}

// GraphQL sends a JSON GraphQL request through the application's HTTP handler.
func GraphQL(t *testing.T, handler http.Handler, query string, variables map[string]any, options ...RequestOption) GraphQLResponse {
	t.Helper()

	body := GraphQLBody(t, query, variables)
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(req)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var response GraphQLResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&response))
	return response
}

// GraphQLBody returns the exact JSON payload used by GraphQL. It is useful when
// the payload must be signed before a request is sent.
func GraphQLBody(t *testing.T, query string, variables map[string]any) []byte {
	t.Helper()
	body, err := json.Marshal(GraphQLRequest{Query: query, Variables: variables})
	require.NoError(t, err)
	return body
}
