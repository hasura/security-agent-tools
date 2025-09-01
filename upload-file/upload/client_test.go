package upload

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_UploadViaReader_Success(t *testing.T) {
	// Create a test server for the upload endpoint
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		// Verify content type
		expectedContentType := "application/json"
		if r.Header.Get("Content-Type") != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, r.Header.Get("Content-Type"))
		}

		// Read and verify content
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		expectedContent := `{"test": "data"}`
		if string(body) != expectedContent {
			t.Errorf("Expected content %s, got %s", expectedContent, string(body))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Upload successful"))
	}))
	defer uploadServer.Close()

	// Create a test server for the GraphQL endpoint
	graphqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method for GraphQL, got %s", r.Method)
		}

		// Return a mock presigned URL response
		response := `{
			"data": {
				"storage_presigned_upload_url": {
					"url": "` + uploadServer.URL + `",
					"expired_at": "` + time.Now().Add(time.Hour).Format(time.RFC3339) + `"
				}
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer graphqlServer.Close()

	// Create client with test servers
	client := NewClient(graphqlServer.URL, "test-api-key")

	// Test content
	testContent := `{"test": "data"}`
	reader := strings.NewReader(testContent)

	// Call the function under test
	err := client.UploadViaReader(context.Background(), reader, "application/json", "test/file.json")

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestClient_UploadViaReader_ReaderError(t *testing.T) {
	// Test case where the reader itself fails
	client := NewClient("https://test.example.com/graphql", "test-key")

	// Create a reader that will fail
	failingReader := &failingReader{err: errors.New("reader error")}

	err := client.UploadViaReader(context.Background(), failingReader, "text/plain", "test/file.txt")

	expectedError := "failed to read content: reader error"
	if err == nil {
		t.Errorf("Expected error containing %q, but got no error", expectedError)
	} else if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing %q, but got: %v", expectedError, err)
	}
}

// failingReader is a mock reader that always returns an error
type failingReader struct {
	err error
}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, f.err
}

func TestClient_UploadViaReader_GraphQLError(t *testing.T) {
	// Create a test server that returns GraphQL error
	graphqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors": [{"message": "Invalid query"}]}`))
	}))
	defer graphqlServer.Close()

	client := NewClient(graphqlServer.URL, "test-key")
	reader := strings.NewReader("test content")

	err := client.UploadViaReader(context.Background(), reader, "text/plain", "test/file.txt")

	if err == nil {
		t.Error("Expected error due to GraphQL failure, but got no error")
	} else if !strings.Contains(err.Error(), "failed to get presigned upload URL") {
		t.Errorf("Expected GraphQL error, but got: %v", err)
	}
}

func TestClient_UploadViaReader_UploadError(t *testing.T) {
	// Create upload server that returns error
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
	}))
	defer uploadServer.Close()

	// Create GraphQL server that returns the upload server URL
	graphqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"storage_presigned_upload_url": {
					"url": "` + uploadServer.URL + `",
					"expired_at": "` + time.Now().Add(time.Hour).Format(time.RFC3339) + `"
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer graphqlServer.Close()

	client := NewClient(graphqlServer.URL, "test-key")
	reader := strings.NewReader("test content")

	err := client.UploadViaReader(context.Background(), reader, "text/plain", "test/file.txt")
	if err == nil {
		t.Error("Expected upload error, but got no error")
	} else if !strings.Contains(err.Error(), "upload failed with status 403") {
		t.Errorf("Expected upload status error, but got: %v", err)
	}
}

func TestClient_UploadViaReader_EmptyContent(t *testing.T) {
	// Test with empty content
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	graphqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"storage_presigned_upload_url": {
					"url": "` + uploadServer.URL + `",
					"expired_at": "` + time.Now().Add(time.Hour).Format(time.RFC3339) + `"
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer graphqlServer.Close()

	client := NewClient(graphqlServer.URL, "test-key")
	reader := strings.NewReader("")

	err := client.UploadViaReader(context.Background(), reader, "text/plain", "test/empty.txt")

	if err != nil {
		t.Errorf("Expected no error for empty content, but got: %v", err)
	}
}
