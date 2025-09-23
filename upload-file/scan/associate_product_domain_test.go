package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/catalog"
	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

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
