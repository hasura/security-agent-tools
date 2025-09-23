package saclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/machinebox/graphql"
)

func TestNewClient(t *testing.T) {
	endpoint := "https://example.com/graphql"
	apiKey := "test-api-key"

	client := NewClient(endpoint, apiKey)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.securityAgentAPIEndpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, client.securityAgentAPIEndpoint)
	}

	if client.securityAgentAPIKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, client.securityAgentAPIKey)
	}

	if client.gqlClient == nil {
		t.Error("GraphQL client should not be nil")
	}

	if client.UploadClient == nil {
		t.Error("Upload client should not be nil")
	}

	expectedTimeout := 5 * time.Minute
	if client.UploadClient.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, client.UploadClient.Timeout)
	}
}

func TestExecuteGQL(t *testing.T) {
	// Create a test server to mock GraphQL endpoint
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
		w.Write([]byte(`{"data": {"test": "success"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	req := graphql.NewRequest(`query { test }`)

	var response map[string]interface{}
	err := client.ExecuteGQL(context.Background(), req, &response)

	if err != nil {
		t.Errorf("ExecuteGQL failed: %v", err)
	}

	if response == nil {
		t.Error("Response should not be nil")
	}
}

func TestUploadFile_FileNotExists(t *testing.T) {
	client := NewClient("https://example.com/graphql", "test-api-key")

	err := client.UploadFile(context.Background(), "/nonexistent/file.txt", "destination.txt")

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("Expected 'file does not exist' error, got: %v", err)
	}
}

func TestUploadFile_Success(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock GraphQL server for presigned URL
	gqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"storage_presigned_upload_url": {
					"url": "%s/upload",
					"expired_at": "%s"
				}
			}
		}`, "http://localhost:8080", time.Now().Add(time.Hour).Format(time.RFC3339))
		w.Write([]byte(response))
	}))
	defer gqlServer.Close()

	// Mock upload server
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type 'text/plain', got '%s'", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		if string(body) != testContent {
			t.Errorf("Expected body '%s', got '%s'", testContent, string(body))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	// Update the GraphQL server to return the upload server URL
	gqlServer.Close()
	gqlServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"storage_presigned_upload_url": {
					"url": "%s",
					"expired_at": "%s"
				}
			}
		}`, uploadServer.URL, time.Now().Add(time.Hour).Format(time.RFC3339))
		w.Write([]byte(response))
	}))
	defer gqlServer.Close()

	client := NewClient(gqlServer.URL, "test-api-key")

	err = client.UploadFile(context.Background(), testFile, "test-destination.txt")

	if err != nil {
		t.Errorf("UploadFile failed: %v", err)
	}
}

func TestPresignedUploadURL_Success(t *testing.T) {
	expectedURL := "https://s3.amazonaws.com/bucket/file.txt?signature=abc123"
	expectedExpiredAt := time.Now().Add(time.Hour)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"storage_presigned_upload_url": {
					"url": "%s",
					"expired_at": "%s"
				}
			}
		}`, expectedURL, expectedExpiredAt.Format(time.RFC3339))
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	url, err := client.presignedUploadURL(context.Background(), "test-file.txt")

	if err != nil {
		t.Errorf("presignedUploadURL failed: %v", err)
	}

	if url != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, url)
	}
}

func TestPresignedUploadURL_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "GraphQL error"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	_, err := client.presignedUploadURL(context.Background(), "test-file.txt")

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}
}

func TestRawUpload_Success(t *testing.T) {
	testContent := "Hello, World!"
	contentType := ContentTypeText
	contentSize := int64(len(testContent))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != string(contentType) {
			t.Errorf("Expected Content-Type '%s', got '%s'", contentType, r.Header.Get("Content-Type"))
		}

		if r.ContentLength != contentSize {
			t.Errorf("Expected Content-Length %d, got %d", contentSize, r.ContentLength)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		if string(body) != testContent {
			t.Errorf("Expected body '%s', got '%s'", testContent, string(body))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("https://example.com/graphql", "test-api-key")
	reader := bytes.NewReader([]byte(testContent))

	err := client.rawUpload(context.Background(), server.URL, contentType, contentSize, reader)

	if err != nil {
		t.Errorf("rawUpload failed: %v", err)
	}
}

func TestRawUpload_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient("https://example.com/graphql", "test-api-key")
	reader := bytes.NewReader([]byte("test content"))

	err := client.rawUpload(context.Background(), server.URL, ContentTypeText, 12, reader)

	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}

	if !strings.Contains(err.Error(), "upload failed with status 500") {
		t.Errorf("Expected status 500 error, got: %v", err)
	}
}

func TestRawUpload_InvalidURL(t *testing.T) {
	client := NewClient("https://example.com/graphql", "test-api-key")
	reader := bytes.NewReader([]byte("test content"))

	err := client.rawUpload(context.Background(), "invalid-url", ContentTypeText, 12, reader)

	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	if !strings.Contains(err.Error(), "failed to upload file") {
		t.Errorf("Expected upload failure error, got: %v", err)
	}
}

func TestRawUpload_NetworkError(t *testing.T) {
	client := NewClient("https://example.com/graphql", "test-api-key")
	reader := bytes.NewReader([]byte("test content"))

	// Use a URL that will cause a network error
	err := client.rawUpload(context.Background(), "http://localhost:1", ContentTypeText, 12, reader)

	if err == nil {
		t.Error("Expected error for network failure")
	}

	if !strings.Contains(err.Error(), "failed to upload file") {
		t.Errorf("Expected upload failure error, got: %v", err)
	}
}

func TestUploadFile_PresignedURLError(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Failed to generate presigned URL"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	err = client.UploadFile(context.Background(), testFile, "test-destination.txt")

	if err == nil {
		t.Error("Expected error when presigned URL generation fails")
	}

	if !strings.Contains(err.Error(), "failed to get presigned upload URL") {
		t.Errorf("Expected presigned URL error, got: %v", err)
	}
}

func TestUploadFile_FileOpenError(t *testing.T) {
	// Test with a non-existent file in a non-existent directory
	nonExistentFile := "/non/existent/path/file.txt"

	client := NewClient("https://example.com/graphql", "test-api-key")

	err := client.UploadFile(context.Background(), nonExistentFile, "test-destination.txt")

	if err == nil {
		t.Error("Expected error when trying to upload a non-existent file")
	}

	if !strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("Expected 'file does not exist' error, got: %v", err)
	}
}

func TestUploadFile_UploadError(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock GraphQL server for presigned URL
	gqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"data": {
				"storage_presigned_upload_url": {
					"url": "%s",
					"expired_at": "%s"
				}
			}
		}`, "http://localhost:1", time.Now().Add(time.Hour).Format(time.RFC3339))
		w.Write([]byte(response))
	}))
	defer gqlServer.Close()

	client := NewClient(gqlServer.URL, "test-api-key")

	err = client.UploadFile(context.Background(), testFile, "test-destination.txt")

	if err == nil {
		t.Error("Expected error when upload fails")
	}
}

func TestUploadFile_ContextCancellation(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient("https://example.com/graphql", "test-api-key")

	err = client.UploadFile(ctx, testFile, "test-destination.txt")

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestUploadFile_DifferentContentTypes(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		content    string
		expectedCT ContentType
	}{
		{
			name:       "JSON file",
			filename:   "data.json",
			content:    `{"key": "value"}`,
			expectedCT: ContentTypeJSON,
		},
		{
			name:       "CSV file",
			filename:   "data.csv",
			content:    "name,age\nJohn,30",
			expectedCT: ContentTypeCSV,
		},
		{
			name:       "XML file",
			filename:   "config.xml",
			content:    "<root><item>value</item></root>",
			expectedCT: ContentTypeXML,
		},
		{
			name:       "Unknown extension",
			filename:   "file.unknown",
			content:    "some content",
			expectedCT: ContentTypeOctetStream,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for testing
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, tt.filename)

			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Mock GraphQL server for presigned URL
			gqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := fmt.Sprintf(`{
					"data": {
						"storage_presigned_upload_url": {
							"url": "%s",
							"expired_at": "%s"
						}
					}
				}`, "UPLOAD_URL_PLACEHOLDER", time.Now().Add(time.Hour).Format(time.RFC3339))
				w.Write([]byte(response))
			}))
			defer gqlServer.Close()

			// Mock upload server to verify content type
			uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Content-Type") != string(tt.expectedCT) {
					t.Errorf("Expected Content-Type '%s', got '%s'", tt.expectedCT, r.Header.Get("Content-Type"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer uploadServer.Close()

			// Test content type detection directly
			contentType := getContentType(testFile)
			if contentType != tt.expectedCT {
				t.Errorf("Expected content type %s, got %s", tt.expectedCT, contentType)
			}

			// Test rawUpload with the correct content type
			client := NewClient("https://example.com/graphql", "test-api-key")
			reader := bytes.NewReader([]byte(tt.content))

			err = client.rawUpload(context.Background(), uploadServer.URL, contentType, int64(len(tt.content)), reader)
			if err != nil {
				t.Errorf("rawUpload failed: %v", err)
			}
		})
	}
}

func TestPresignedUploadURL_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Use a valid timestamp format even if empty URL
		validTime := time.Now().Format(time.RFC3339)
		response := fmt.Sprintf(`{"data": {"storage_presigned_upload_url": {"url": "", "expired_at": "%s"}}}`, validTime)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	url, err := client.presignedUploadURL(context.Background(), "test-file.txt")

	if err != nil {
		t.Errorf("presignedUploadURL failed: %v", err)
	}

	if url != "" {
		t.Errorf("Expected empty URL, got %s", url)
	}
}

func TestPresignedUploadURL_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"invalid": "response"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	url, err := client.presignedUploadURL(context.Background(), "test-file.txt")

	if err != nil {
		t.Errorf("presignedUploadURL failed: %v", err)
	}

	if url != "" {
		t.Errorf("Expected empty URL for malformed response, got %s", url)
	}
}

func TestRawUpload_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expectErr  bool
	}{
		{"Success 200", http.StatusOK, false},
		{"Success 201", http.StatusCreated, false},
		{"Success 204", http.StatusNoContent, false},
		{"Client Error 400", http.StatusBadRequest, true},
		{"Client Error 403", http.StatusForbidden, true},
		{"Client Error 404", http.StatusNotFound, true},
		{"Server Error 500", http.StatusInternalServerError, true},
		{"Server Error 503", http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.expectErr {
					w.Write([]byte("Error response"))
				}
			}))
			defer server.Close()

			client := NewClient("https://example.com/graphql", "test-api-key")
			reader := bytes.NewReader([]byte("test content"))

			err := client.rawUpload(context.Background(), server.URL, ContentTypeText, 12, reader)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for status code %d", tt.statusCode)
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for status code %d: %v", tt.statusCode, err)
			}
		})
	}
}
