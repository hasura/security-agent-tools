package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/catalog"
	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestScan_AssociateProductDomains_Success(t *testing.T) {
	expectedProductDomain := "hasura-v2-cloud-control-plane"

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

	productDomain, err := scan.AssociateProductDomains(context.Background())

	if err != nil {
		t.Errorf("AssociateProductDomains failed: %v", err)
	}

	if productDomain != expectedProductDomain {
		t.Errorf("Expected product domain %s, got %s", expectedProductDomain, productDomain)
	}
}

func TestScan_AssociateProductDomains_EmptyProductDomain(t *testing.T) {
	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{}, // No product_domain tag
		client: client,
	}

	productDomain, err := scan.AssociateProductDomains(context.Background())

	if err != nil {
		t.Errorf("AssociateProductDomains failed: %v", err)
	}

	if productDomain != "" {
		t.Errorf("Expected empty product domain, got %s", productDomain)
	}
}

func TestScan_AssociateProductDomains_GraphQLError(t *testing.T) {
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
			"product_domain": "hasura-ddn-control-plane",
		},
		client: client,
	}

	productDomain, err := scan.AssociateProductDomains(context.Background())

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if productDomain != "hasura-ddn-control-plane" {
		t.Errorf("Expected product domain to be returned even on error, got %s", productDomain)
	}
}

func TestScan_AssociateProductDomains_MultipleDomainsSuccess(t *testing.T) {
	expectedProductDomains := "hasura-v2-cloud-control-plane, hasura-v2-cloud-data-plane, hasura-ddn-control-plane"
	callCount := 0

	// Mock GraphQL server that expects 3 calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
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
			"product_domain": expectedProductDomains,
		},
		client: client,
	}

	productDomain, err := scan.AssociateProductDomains(context.Background())

	if err != nil {
		t.Errorf("AssociateProductDomains failed: %v", err)
	}

	if productDomain != expectedProductDomains {
		t.Errorf("Expected product domain %s, got %s", expectedProductDomains, productDomain)
	}

	// Should have made 3 GraphQL calls (one for each domain)
	if callCount != 3 {
		t.Errorf("Expected 3 GraphQL calls, got %d", callCount)
	}
}

func TestScan_AssociateProductDomains_MultipleDomainsPartialError(t *testing.T) {
	expectedProductDomains := "hasura-v2-cloud-control-plane, hasura-v2-cloud-data-plane, hasura-ddn-control-plane"
	callCount := 0

	// Mock GraphQL server that fails on the second call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if callCount == 2 {
			// Fail on second call (hasura-v2-cloud-data-plane)
			w.Write([]byte(`{"errors": [{"message": "Database error"}]}`))
		} else {
			w.Write([]byte(`{
				"data": {
					"insert_vulnerability_reports_by_product_domains": {
						"returning": [{
							"id": "assoc-id-456"
						}]
					}
				}
			}`))
		}
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID: "scan-id-123",
		Tags: map[string]string{
			"product_domain": expectedProductDomains,
		},
		client: client,
	}

	productDomain, err := scan.AssociateProductDomains(context.Background())

	// Should return an error due to the failed second call
	if err == nil {
		t.Error("Expected error due to partial failure")
	}

	// Should still return the original product domain string
	if productDomain != expectedProductDomains {
		t.Errorf("Expected product domain %s, got %s", expectedProductDomains, productDomain)
	}

	// Should have made 3 GraphQL calls (one for each domain)
	if callCount != 3 {
		t.Errorf("Expected 3 GraphQL calls, got %d", callCount)
	}
}

func TestProductDomains_Success(t *testing.T) {
	expectedDomains := []string{"hasura-v2-cloud-control-plane", "hasura-v2-cloud-data-plane", "hasura-ddn-control-plane"}

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	domains, err := catalog.ProductDomains(context.Background(), client)

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
	domains, err := catalog.ProductDomains(context.Background(), client)

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
	domains, err := catalog.ProductDomains(context.Background(), client)

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if domains != nil {
		t.Error("Expected domains to be nil when error occurs")
	}
}
