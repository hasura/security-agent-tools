package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestScan_AssociateImageName_Success(t *testing.T) {
	expectedImageName := "nginx:latest"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_image_name": {
					"returning": [{
						"id": "assoc-id-123",
						"scan_id": "scan-id-123"
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
			"image_name": expectedImageName,
		},
		client: client,
	}

	imageName, err := scan.AssociateImageName(context.Background())

	if err != nil {
		t.Errorf("AssociateImageName failed: %v", err)
	}

	if imageName != expectedImageName {
		t.Errorf("Expected image name %s, got %s", expectedImageName, imageName)
	}
}

func TestScan_AssociateImageName_EmptyImageName(t *testing.T) {
	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{}, // No image_name tag
		client: client,
	}

	imageName, err := scan.AssociateImageName(context.Background())

	if err != nil {
		t.Errorf("AssociateImageName failed: %v", err)
	}

	if imageName != "" {
		t.Errorf("Expected empty image name, got %s", imageName)
	}
}

func TestScan_AssociateImageName_GraphQLError(t *testing.T) {
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
			"image_name": "nginx:latest",
		},
		client: client,
	}

	imageName, err := scan.AssociateImageName(context.Background())

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if imageName != "nginx:latest" {
		t.Errorf("Expected image name to be returned even on error, got %s", imageName)
	}
}
