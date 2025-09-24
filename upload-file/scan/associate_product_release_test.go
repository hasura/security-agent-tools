package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestScan_AssociateProductRelease_Success(t *testing.T) {
	expectedProductRelease := "v1.2.3"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_product_release": {
					"returning": [{
						"id": "product-release-assoc-id-123"
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
			"product_release": expectedProductRelease,
		},
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err != nil {
		t.Errorf("AssociateProductRelease failed: %v", err)
	}

	if productRelease != expectedProductRelease {
		t.Errorf("Expected product release %s, got %s", expectedProductRelease, productRelease)
	}
}

func TestScan_AssociateProductRelease_EmptyProductRelease(t *testing.T) {
	client := saclient.NewClient("http://localhost", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{}, // No product_release tag
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err != nil {
		t.Errorf("AssociateProductRelease failed: %v", err)
	}

	if productRelease != "" {
		t.Errorf("Expected empty product release, got %s", productRelease)
	}
}

func TestScan_AssociateProductRelease_EmptyProductReleaseValue(t *testing.T) {
	client := saclient.NewClient("http://localhost", "test-api-key")
	scan := &Scan{
		ID: "scan-id-123",
		Tags: map[string]string{
			"product_release": "", // Empty product_release value
		},
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err != nil {
		t.Errorf("AssociateProductRelease failed: %v", err)
	}

	if productRelease != "" {
		t.Errorf("Expected empty product release, got %s", productRelease)
	}
}

func TestScan_AssociateProductRelease_GraphQLError(t *testing.T) {
	expectedProductRelease := "v2.0.0"

	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database connection failed"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID: "scan-id-123",
		Tags: map[string]string{
			"product_release": expectedProductRelease,
		},
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if productRelease != expectedProductRelease {
		t.Errorf("Expected product release %s even on error, got %s", expectedProductRelease, productRelease)
	}
}

func TestScan_AssociateProductRelease_NetworkError(t *testing.T) {
	expectedProductRelease := "v3.1.0"

	// Mock GraphQL server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID: "scan-id-123",
		Tags: map[string]string{
			"product_release": expectedProductRelease,
		},
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err == nil {
		t.Error("Expected error for network error response")
	}

	if productRelease != expectedProductRelease {
		t.Errorf("Expected product release %s even on error, got %s", expectedProductRelease, productRelease)
	}
}

func TestScan_AssociateProductRelease_WithSpecialCharacters(t *testing.T) {
	expectedProductRelease := "v1.0.0-beta.1+build.123"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_product_release": {
					"returning": [{
						"id": "product-release-assoc-id-456"
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
			"product_release": expectedProductRelease,
		},
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err != nil {
		t.Errorf("AssociateProductRelease failed: %v", err)
	}

	if productRelease != expectedProductRelease {
		t.Errorf("Expected product release %s, got %s", expectedProductRelease, productRelease)
	}
}

func TestScan_AssociateProductRelease_WithOtherTags(t *testing.T) {
	expectedProductRelease := "v4.5.6"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_product_release": {
					"returning": [{
						"id": "product-release-assoc-id-789"
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
			"product_release": expectedProductRelease,
			"service":         "api-gateway",
			"environment":     "production",
			"team":            "backend",
		},
		client: client,
	}

	productRelease, err := scan.AssociateProductRelease(context.Background())

	if err != nil {
		t.Errorf("AssociateProductRelease failed: %v", err)
	}

	if productRelease != expectedProductRelease {
		t.Errorf("Expected product release %s, got %s", expectedProductRelease, productRelease)
	}
}
