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
