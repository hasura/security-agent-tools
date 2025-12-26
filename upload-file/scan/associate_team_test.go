package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestScan_AssociateTeam_Success(t *testing.T) {
	expectedTeam := "backend-team"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_team": {
					"returning": [{
						"id": "team-assoc-id-456"
					}]
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID: "scan-id-123",
		Tags: map[string]string{
			"team": expectedTeam,
		},
		client: client,
	}

	team, err := scan.AssociateTeam(context.Background())

	if err != nil {
		t.Errorf("AssociateTeam failed: %v", err)
	}

	if team != expectedTeam {
		t.Errorf("Expected team %s, got %s", expectedTeam, team)
	}
}

func TestScan_AssociateTeam_EmptyTeam(t *testing.T) {
	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{}, // No team tag
		client: client,
	}

	team, err := scan.AssociateTeam(context.Background())

	if err != nil {
		t.Errorf("AssociateTeam failed: %v", err)
	}

	if team != "" {
		t.Errorf("Expected empty team, got %s", team)
	}
}

func TestScan_AssociateTeam_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database error"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID: "scan-id-123",
		Tags: map[string]string{
			"team": "backend-team",
		},
		client: client,
	}

	team, err := scan.AssociateTeam(context.Background())

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if team != "backend-team" {
		t.Errorf("Expected team to be returned even on error, got %s", team)
	}
}
