package scan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestNew_Success(t *testing.T) {
	expectedID := "test-scan-id-123"
	expectedTags := map[string]string{
		"environment": "test",
		"service":     "api",
	}

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Authorization") != "test-api-key" {
			t.Errorf("Expected Authorization header 'test-api-key', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Hasura-Auth-Mode") != "ci-auth" {
			t.Errorf("Expected X-Hasura-Auth-Mode header 'ci-auth', got '%s'", r.Header.Get("X-Hasura-Auth-Mode"))
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"insert_vulnerability_reports_scans": {
					"returning": [{
						"id": "%s",
						"tags": {
							"environment": "test",
							"service": "api"
						}
					}]
				}
			}
		}`, expectedID)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan, err := New(context.Background(), client, expectedTags)

	if err != nil {
		t.Errorf("New failed: %v", err)
	}

	if scan == nil {
		t.Fatal("Expected scan to not be nil")
	}

	if scan.ID != expectedID {
		t.Errorf("Expected scan ID %s, got %s", expectedID, scan.ID)
	}

	if len(scan.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(scan.Tags))
	}

	for key, expectedValue := range expectedTags {
		if actualValue, exists := scan.Tags[key]; !exists {
			t.Errorf("Expected tag %s to exist", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected tag %s to have value %s, got %s", key, expectedValue, actualValue)
		}
	}

	if scan.client == nil {
		t.Error("Expected scan client to not be nil")
	}
}

func TestNew_WithNilTags(t *testing.T) {
	expectedID := "test-scan-id-456"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"insert_vulnerability_reports_scans": {
					"returning": [{
						"id": "%s",
						"tags": {}
					}]
				}
			}
		}`, expectedID)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan, err := New(context.Background(), client, nil)

	if err != nil {
		t.Errorf("New failed: %v", err)
	}

	if scan == nil {
		t.Fatal("Expected scan to not be nil")
	}

	if scan.ID != expectedID {
		t.Errorf("Expected scan ID %s, got %s", expectedID, scan.ID)
	}

	if scan.Tags == nil {
		t.Error("Expected scan tags to not be nil")
	}

	if len(scan.Tags) != 0 {
		t.Errorf("Expected empty tags map, got %d tags", len(scan.Tags))
	}
}

func TestNew_WithEmptyTags(t *testing.T) {
	expectedID := "test-scan-id-789"
	emptyTags := make(map[string]string)

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"insert_vulnerability_reports_scans": {
					"returning": [{
						"id": "%s",
						"tags": {}
					}]
				}
			}
		}`, expectedID)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan, err := New(context.Background(), client, emptyTags)

	if err != nil {
		t.Errorf("New failed: %v", err)
	}

	if scan == nil {
		t.Fatal("Expected scan to not be nil")
	}

	if scan.ID != expectedID {
		t.Errorf("Expected scan ID %s, got %s", expectedID, scan.ID)
	}

	if len(scan.Tags) != 0 {
		t.Errorf("Expected empty tags map, got %d tags", len(scan.Tags))
	}
}

func TestNew_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database connection failed"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	tags := map[string]string{"test": "value"}

	scan, err := New(context.Background(), client, tags)

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if scan != nil {
		t.Error("Expected scan to be nil when error occurs")
	}
}

func TestNew_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	client := saclient.NewClient("http://localhost:1", "test-api-key")
	tags := map[string]string{"test": "value"}

	scan, err := New(context.Background(), client, tags)

	if err == nil {
		t.Error("Expected error for network failure")
	}

	if scan != nil {
		t.Error("Expected scan to be nil when error occurs")
	}
}

func TestNew_ContextCancellation(t *testing.T) {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	tags := map[string]string{"test": "value"}

	scan, err := New(ctx, client, tags)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}

	if scan != nil {
		t.Error("Expected scan to be nil when context is cancelled")
	}
}

func TestNew_EmptyResponse(t *testing.T) {
	// Mock GraphQL server that returns empty returning array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_scans": {
					"returning": []
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	tags := map[string]string{"test": "value"}

	// This should panic due to index out of bounds, but let's test it
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for empty returning array")
		}
	}()

	New(context.Background(), client, tags)
}

func TestNew_MalformedResponse(t *testing.T) {
	// Mock GraphQL server that returns malformed JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"invalid": "structure"}}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	tags := map[string]string{"test": "value"}

	// This should panic due to index out of bounds, but let's test it
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for malformed response structure")
		}
	}()

	New(context.Background(), client, tags)
}

func TestScan_AssociateProductDomain_Success(t *testing.T) {
	expectedProductDomain := "ecommerce"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_by_product_domains": {
					"returning": [{
						"id": "assoc-id-456"
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
			"product_domain": expectedProductDomain,
		},
		client: client,
	}

	productDomain, err := scan.AssociateProductDomain(context.Background())

	if err != nil {
		t.Errorf("AssociateProductDomain failed: %v", err)
	}

	if productDomain != expectedProductDomain {
		t.Errorf("Expected product domain %s, got %s", expectedProductDomain, productDomain)
	}
}

func TestScan_AssociateProductDomain_EmptyProductDomain(t *testing.T) {
	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{}, // No product_domain tag
		client: client,
	}

	productDomain, err := scan.AssociateProductDomain(context.Background())

	if err != nil {
		t.Errorf("AssociateProductDomain failed: %v", err)
	}

	if productDomain != "" {
		t.Errorf("Expected empty product domain, got %s", productDomain)
	}
}

func TestScan_AssociateProductDomain_GraphQLError(t *testing.T) {
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
			"product_domain": "ecommerce",
		},
		client: client,
	}

	productDomain, err := scan.AssociateProductDomain(context.Background())

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if productDomain != "ecommerce" {
		t.Errorf("Expected product domain to be returned even on error, got %s", productDomain)
	}
}

func TestProductDomains_Success(t *testing.T) {
	expectedDomains := []string{"ecommerce", "analytics", "security"}

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"vulnerability_reports_product_domains": [
					{"code": "ecommerce"},
					{"code": "analytics"},
					{"code": "security"}
				]
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	domains, err := ProductDomains(context.Background(), client)

	if err != nil {
		t.Errorf("ProductDomains failed: %v", err)
	}

	if len(domains) != len(expectedDomains) {
		t.Errorf("Expected %d domains, got %d", len(expectedDomains), len(domains))
	}

	for i, expectedDomain := range expectedDomains {
		if i >= len(domains) || domains[i] != expectedDomain {
			t.Errorf("Expected domain %s at index %d, got %s", expectedDomain, i, domains[i])
		}
	}
}

func TestProductDomains_EmptyResponse(t *testing.T) {
	// Mock GraphQL server that returns empty array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"vulnerability_reports_product_domains": []
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	domains, err := ProductDomains(context.Background(), client)

	if err != nil {
		t.Errorf("ProductDomains failed: %v", err)
	}

	if len(domains) != 0 {
		t.Errorf("Expected empty domains array, got %d domains", len(domains))
	}
}

func TestProductDomains_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database error"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	domains, err := ProductDomains(context.Background(), client)

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if domains != nil {
		t.Error("Expected domains to be nil when error occurs")
	}
}

func TestScan_AssociateScanReport_Success(t *testing.T) {
	expectedReportPath := "s3://bucket/reports/scan-123.json"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_scan_reports": {
					"returning": [{
						"id": "report-assoc-id-789"
					}]
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	err := scan.AssociateScanReport(context.Background(), expectedReportPath)

	if err != nil {
		t.Errorf("AssociateScanReport failed: %v", err)
	}
}

func TestScan_AssociateScanReport_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database error"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	err := scan.AssociateScanReport(context.Background(), "s3://bucket/reports/scan-123.json")

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}
}

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
