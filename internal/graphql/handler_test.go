package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/graphql/resolver"
	"github.com/soumajitgh/mobicode/internal/service/task"

	"go.uber.org/zap"
)

type fakeTaskService struct{}

// Create returns a deterministic Task for GraphQL handler tests.
func (fakeTaskService) Create(context.Context, task.CreateInput) (*task.Task, error) {
	return &task.Task{ID: 1, Title: "first task", Status: "pending", CreatedAt: time.Unix(0, 0), UpdatedAt: time.Unix(0, 0)}, nil
}

// Get returns no Task because retrieval is outside this test's scope.
func (fakeTaskService) Get(context.Context, uint) (*task.Task, error) { return nil, nil }

// List returns an empty Task page because listing is outside this test's scope.
func (fakeTaskService) List(context.Context, task.ListInput) (*task.TaskConnection, error) {
	return &task.TaskConnection{}, nil
}

// Update returns no Task because updating is outside this test's scope.
func (fakeTaskService) Update(context.Context, uint, task.UpdateInput) (*task.Task, error) {
	return nil, nil
}

// Delete succeeds because deletion is outside this test's scope.
func (fakeTaskService) Delete(context.Context, uint) error { return nil }

// TestHandlerServesGraphQLOperations verifies a root query and mutation execute through gqlgen.
func TestHandlerServesGraphQLOperations(t *testing.T) {
	handler := NewHandler(config.Config{Server: config.ServerConfig{GraphQLComplexity: 250}}, resolver.New(fakeTaskService{}), zap.NewNop())
	for _, testCase := range []struct{ name, query, want string }{
		{name: "ping", query: `{ "query": "query { ping }" }`, want: `"ping":"pong"`},
		{name: "create task", query: `{ "query": "mutation { createTask(input: { title: \"first task\" }) { id title status } }" }`, want: `"title":"first task"`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewBufferString(testCase.query))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			var body map[string]any
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			encoded, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("encode response: %v", err)
			}
			if !bytes.Contains(encoded, []byte(testCase.want)) {
				t.Fatalf("response = %s, want %s", encoded, testCase.want)
			}
		})
	}
}
