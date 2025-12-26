package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestTeams_Success(t *testing.T) {
	expectedTeams := []string{"backend-team", "frontend-team", "devops-team"}

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request contains the expected query
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		if !strings.Contains(string(body), "team_catalog_teams") {
			t.Error("Expected query to contain team_catalog_teams")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"team_catalog_teams": [
					{"name": "backend-team"},
					{"name": "frontend-team"},
					{"name": "devops-team"}
				]
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	teams, err := Teams(context.Background(), client)

	if err != nil {
		t.Errorf("Teams failed: %v", err)
	}

	if len(teams) != len(expectedTeams) {
		t.Errorf("Expected %d teams, got %d", len(expectedTeams), len(teams))
	}

	for i, expectedTeam := range expectedTeams {
		if i >= len(teams) || teams[i] != expectedTeam {
			t.Errorf("Expected team %s at index %d, got %s", expectedTeam, i, teams[i])
		}
	}
}

func TestTeams_EmptyResponse(t *testing.T) {
	// Mock GraphQL server that returns empty array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"team_catalog_teams": []
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	teams, err := Teams(context.Background(), client)

	if err != nil {
		t.Errorf("Teams failed: %v", err)
	}

	if len(teams) != 0 {
		t.Errorf("Expected empty teams array, got %d teams", len(teams))
	}
}

func TestTeams_SingleTeam(t *testing.T) {
	expectedTeam := "security-team"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"team_catalog_teams": [
					{"name": "security-team"}
				]
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	teams, err := Teams(context.Background(), client)

	if err != nil {
		t.Errorf("Teams failed: %v", err)
	}

	if len(teams) != 1 {
		t.Errorf("Expected 1 team, got %d", len(teams))
	}

	if teams[0] != expectedTeam {
		t.Errorf("Expected team %s, got %s", expectedTeam, teams[0])
	}
}

func TestTeams_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database error"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	teams, err := Teams(context.Background(), client)

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if teams != nil {
		t.Error("Expected teams to be nil when error occurs")
	}
}

func TestTeams_MalformedJSON(t *testing.T) {
	// Mock GraphQL server that returns malformed JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"team_catalog_teams": [{"name": "incomplete"`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	teams, err := Teams(context.Background(), client)

	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}

	if teams != nil {
		t.Error("Expected teams to be nil when JSON parsing fails")
	}
}
