package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestProductDomains_Success(t *testing.T) {
	expectedDomains := []string{"hasura-v2-cloud-control-plane", "hasura-v2-cloud-data-plane", "hasura-ddn-control-plane"}

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify headers
		if r.Header.Get("Authorization") != "test-api-key" {
			t.Errorf("Expected Authorization header 'test-api-key', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Hasura-Auth-Mode") != "ci-auth" {
			t.Errorf("Expected X-Hasura-Auth-Mode header 'ci-auth', got '%s'", r.Header.Get("X-Hasura-Auth-Mode"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"vulnerability_reports_product_domains": [
					{"code": "hasura-v2-cloud-control-plane"},
					{"code": "hasura-v2-cloud-data-plane"},
					{"code": "hasura-ddn-control-plane"}
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

	// Verify that we get nil (zero value for []string when no elements are appended)
	if domains != nil {
		t.Error("Expected nil slice when no domains are returned, got non-nil")
	}
}

func TestProductDomains_SingleDomain(t *testing.T) {
	expectedDomain := "hasura-ddn-ci"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"vulnerability_reports_product_domains": [
					{"code": "hasura-ddn-ci"}
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

	if len(domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(domains))
	}

	if domains[0] != expectedDomain {
		t.Errorf("Expected domain %s, got %s", expectedDomain, domains[0])
	}
}

func TestProductDomains_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database connection failed"}]}`))
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

	if !strings.Contains(err.Error(), "Database connection failed") {
		t.Errorf("Expected error message to contain 'Database connection failed', got: %v", err)
	}
}

func TestProductDomains_HTTPError(t *testing.T) {
	// Mock GraphQL server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	domains, err := ProductDomains(context.Background(), client)

	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}

	if domains != nil {
		t.Error("Expected domains to be nil when error occurs")
	}
}

func TestProductDomains_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	client := saclient.NewClient("http://localhost:1", "test-api-key")
	domains, err := ProductDomains(context.Background(), client)

	if err == nil {
		t.Error("Expected error for network failure")
	}

	if domains != nil {
		t.Error("Expected domains to be nil when error occurs")
	}
}

func TestProductDomains_ContextCancellation(t *testing.T) {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	domains, err := ProductDomains(ctx, client)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if domains != nil {
		t.Error("Expected domains to be nil when context is cancelled")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestProductDomains_MalformedJSON(t *testing.T) {
	// Mock GraphQL server that returns malformed JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"vulnerability_reports_product_domains": [{"code": "incomplete"`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	domains, err := ProductDomains(context.Background(), client)

	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}

	if domains != nil {
		t.Error("Expected domains to be nil when JSON parsing fails")
	}
}

func TestProductDomains_NilClient(t *testing.T) {
	// Test with nil client - this should panic or return an error
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling ProductDomains with nil client")
		}
	}()

	ProductDomains(context.Background(), nil)
}

func TestProductDomains_EmptyCodeFields(t *testing.T) {
	// Mock GraphQL server that returns domains with empty code fields
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"vulnerability_reports_product_domains": [
					{"code": ""},
					{"code": "promptql-control-plane"},
					{"code": ""}
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

	if len(domains) != 3 {
		t.Errorf("Expected 3 domains (including empty ones), got %d", len(domains))
	}

	expectedDomains := []string{"", "promptql-control-plane", ""}
	for i, expectedDomain := range expectedDomains {
		if i >= len(domains) || domains[i] != expectedDomain {
			t.Errorf("Expected domain '%s' at index %d, got '%s'", expectedDomain, i, domains[i])
		}
	}
}
