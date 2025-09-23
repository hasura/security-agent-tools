package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestScan_AssociateServiceName_Success(t *testing.T) {
	expectedServiceName := "api-gateway"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_service_name": {
					"returning": [{
						"id": "service-assoc-id-101"
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
			"service": expectedServiceName,
		},
		client: client,
	}

	serviceName, err := scan.AssociateServiceName(context.Background())

	if err != nil {
		t.Errorf("AssociateServiceName failed: %v", err)
	}

	if serviceName != expectedServiceName {
		t.Errorf("Expected service name %s, got %s", expectedServiceName, serviceName)
	}
}

func TestScan_AssociateServiceName_EmptyServiceName(t *testing.T) {
	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{}, // No service tag
		client: client,
	}

	serviceName, err := scan.AssociateServiceName(context.Background())

	if err != nil {
		t.Errorf("AssociateServiceName failed: %v", err)
	}

	if serviceName != "" {
		t.Errorf("Expected empty service name, got %s", serviceName)
	}
}

func TestScan_AssociateServiceName_GraphQLError(t *testing.T) {
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
			"service": "api-gateway",
		},
		client: client,
	}

	serviceName, err := scan.AssociateServiceName(context.Background())

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if serviceName != "api-gateway" {
		t.Errorf("Expected service name to be returned even on error, got %s", serviceName)
	}
}
