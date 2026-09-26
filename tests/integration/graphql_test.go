package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/soumajitgh/mobicode/internal/auth"
)

func TestGraphQLAvailabilityFollowsOnboardingState(t *testing.T) {
	testApp := newTestApplication(t)

	response := testApp.postGraphQL(`{"query":"{ health { status } }"}`)
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GraphQL before onboarding: status = %d", response.StatusCode)
	}
	closeBody(t, response)

	service := auth.Service{Users: testApp.store.Users}
	if _, err := service.CreateInitialUser(context.Background(), "owner@example.com", "validpassword123"); err != nil {
		t.Fatal(err)
	}

	response = testApp.postGraphQL(`{"query":"{ health { status } }"}`)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GraphQL after onboarding: status = %d", response.StatusCode)
	}
	var payload struct {
		Data struct {
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
		} `json:"data"`
		Errors json.RawMessage `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	closeBody(t, response)
	if payload.Data.Health.Status != "ok" || len(payload.Errors) != 0 {
		t.Fatalf("unexpected GraphQL response: %+v", payload)
	}

	response = testApp.get("/mobile/graphql")
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GraphQL GET: status = %d", response.StatusCode)
	}
	closeBody(t, response)
}
