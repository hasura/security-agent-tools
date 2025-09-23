package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestGithubRepoID_NoEnvironmentVariable(t *testing.T) {
	// Ensure GITHUB_REPOSITORY is not set
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Unsetenv("GITHUB_REPOSITORY")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		}
	}()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	repoID, err := GithubRepoID(context.Background(), client)

	if err != nil {
		t.Errorf("GithubRepoID failed: %v", err)
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when GITHUB_REPOSITORY is not set, got %s", repoID)
	}
}

func TestGithubRepoID_InvalidRepositoryFormat(t *testing.T) {
	testCases := []struct {
		name           string
		repoName       string
		expectsGraphQL bool // whether this format will trigger GraphQL calls
	}{
		{"single part", "invalid-repo", false},
		{"too many parts", "org/repo/extra", false},
		{"empty string", "", false},
		{"only slash", "/", true},         // splits to ["", ""] - 2 parts, will make GraphQL calls
		{"slash at start", "/repo", true}, // splits to ["", "repo"] - 2 parts, will make GraphQL calls
		{"slash at end", "org/", true},    // splits to ["org", ""] - 2 parts, will make GraphQL calls
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalValue := os.Getenv("GITHUB_REPOSITORY")
			os.Setenv("GITHUB_REPOSITORY", tc.repoName)
			defer func() {
				if originalValue != "" {
					os.Setenv("GITHUB_REPOSITORY", originalValue)
				} else {
					os.Unsetenv("GITHUB_REPOSITORY")
				}
			}()

			if tc.expectsGraphQL {
				// Mock GraphQL server for cases that will make GraphQL calls
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"data": {
							"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": null
						}
					}`))
				}))
				defer server.Close()

				client := saclient.NewClient(server.URL, "test-api-key")
				repoID, err := GithubRepoID(context.Background(), client)

				// These cases will try to make GraphQL calls but should handle empty org/repo names gracefully
				// The behavior depends on how the GraphQL endpoint handles empty strings
				if err != nil {
					// If there's an error, that's acceptable for invalid input
					t.Logf("GithubRepoID returned error for invalid format %s: %v", tc.repoName, err)
				}

				if repoID != "" {
					t.Errorf("Expected empty repo ID for invalid format %s, got %s", tc.repoName, repoID)
				}
			} else {
				client := saclient.NewClient("https://example.com/graphql", "test-api-key")
				repoID, err := GithubRepoID(context.Background(), client)

				if err != nil {
					t.Errorf("GithubRepoID failed: %v", err)
				}

				if repoID != "" {
					t.Errorf("Expected empty repo ID for invalid format %s, got %s", tc.repoName, repoID)
				}
			}
		})
	}
}

func TestGithubRepoID_ExistingRepo(t *testing.T) {
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Setenv("GITHUB_REPOSITORY", "hasura/graphql-engine")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	expectedID := "test-repo-id-123"

	// Mock GraphQL server that returns existing repo
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
				"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": {
					"id": "` + expectedID + `"
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := GithubRepoID(context.Background(), client)

	if err != nil {
		t.Errorf("GithubRepoID failed: %v", err)
	}

	if repoID != expectedID {
		t.Errorf("Expected repo ID %s, got %s", expectedID, repoID)
	}
}

func TestGithubRepoID_NewRepo(t *testing.T) {
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Setenv("GITHUB_REPOSITORY", "hasura/new-repo")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	expectedID := "new-repo-id-456"
	callCount := 0

	// Mock GraphQL server that first returns not found, then creates new repo
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if callCount == 1 {
			// First call - repo not found
			w.Write([]byte(`{
				"data": {
					"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": null
				}
			}`))
		} else {
			// Second call - create new repo
			w.Write([]byte(`{
				"data": {
					"insert_service_catalog_github_repos": {
						"id": "` + expectedID + `"
					}
				}
			}`))
		}
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := GithubRepoID(context.Background(), client)

	if err != nil {
		t.Errorf("GithubRepoID failed: %v", err)
	}

	if repoID != expectedID {
		t.Errorf("Expected repo ID %s, got %s", expectedID, repoID)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 GraphQL calls (get + add), got %d", callCount)
	}
}

func TestGetGitHubRepo_Success(t *testing.T) {
	expectedID := "existing-repo-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": {
					"id": "` + expectedID + `"
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := getGitHubRepo(context.Background(), client, "hasura", "graphql-engine")

	if err != nil {
		t.Errorf("getGitHubRepo failed: %v", err)
	}

	if repoID != expectedID {
		t.Errorf("Expected repo ID %s, got %s", expectedID, repoID)
	}
}

func TestGetGitHubRepo_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": null
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := getGitHubRepo(context.Background(), client, "hasura", "nonexistent-repo")

	if err != ErrGitHubRepoNotFound {
		t.Errorf("Expected ErrGitHubRepoNotFound, got %v", err)
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when not found, got %s", repoID)
	}
}

func TestGetGitHubRepo_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database connection failed"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := getGitHubRepo(context.Background(), client, "hasura", "graphql-engine")

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}

	if !strings.Contains(err.Error(), "Database connection failed") {
		t.Errorf("Expected error message to contain 'Database connection failed', got: %v", err)
	}
}

func TestAddGitHubRepo_Success(t *testing.T) {
	expectedID := "new-repo-id-789"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_service_catalog_github_repos": {
					"id": "` + expectedID + `"
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := addGitHubRepo(context.Background(), client, "hasura", "new-repo")

	if err != nil {
		t.Errorf("addGitHubRepo failed: %v", err)
	}

	if repoID != expectedID {
		t.Errorf("Expected repo ID %s, got %s", expectedID, repoID)
	}
}

func TestAddGitHubRepo_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Insertion failed"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := addGitHubRepo(context.Background(), client, "hasura", "new-repo")

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}

	if !strings.Contains(err.Error(), "Insertion failed") {
		t.Errorf("Expected error message to contain 'Insertion failed', got: %v", err)
	}
}

func TestGithubRepoID_HTTPError(t *testing.T) {
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	// Mock GraphQL server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := GithubRepoID(context.Background(), client)

	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}
}

func TestGithubRepoID_NetworkError(t *testing.T) {
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	// Use an invalid URL to simulate network error
	client := saclient.NewClient("http://localhost:1", "test-api-key")
	repoID, err := GithubRepoID(context.Background(), client)

	if err == nil {
		t.Error("Expected error for network failure")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}
}

func TestGithubRepoID_ContextCancellation(t *testing.T) {
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	repoID, err := GithubRepoID(ctx, client)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when context is cancelled, got %s", repoID)
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestGetGitHubRepo_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := getGitHubRepo(context.Background(), client, "hasura", "test-repo")

	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}
}

func TestGetGitHubRepo_NetworkError(t *testing.T) {
	client := saclient.NewClient("http://localhost:1", "test-api-key")
	repoID, err := getGitHubRepo(context.Background(), client, "hasura", "test-repo")

	if err == nil {
		t.Error("Expected error for network failure")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}
}

func TestGetGitHubRepo_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": {"id": "incomplete"`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := getGitHubRepo(context.Background(), client, "hasura", "test-repo")

	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when JSON parsing fails, got %s", repoID)
	}
}

func TestAddGitHubRepo_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := addGitHubRepo(context.Background(), client, "hasura", "new-repo")

	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}
}

func TestAddGitHubRepo_NetworkError(t *testing.T) {
	client := saclient.NewClient("http://localhost:1", "test-api-key")
	repoID, err := addGitHubRepo(context.Background(), client, "hasura", "new-repo")

	if err == nil {
		t.Error("Expected error for network failure")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when error occurs, got %s", repoID)
	}
}

func TestAddGitHubRepo_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"insert_service_catalog_github_repos": {"id": "incomplete"`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	repoID, err := addGitHubRepo(context.Background(), client, "hasura", "new-repo")

	if err == nil {
		t.Error("Expected error for malformed JSON response")
	}

	if repoID != "" {
		t.Errorf("Expected empty repo ID when JSON parsing fails, got %s", repoID)
	}
}

func TestGithubRepoID_NilClient(t *testing.T) {
	originalValue := os.Getenv("GITHUB_REPOSITORY")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")
	defer func() {
		if originalValue != "" {
			os.Setenv("GITHUB_REPOSITORY", originalValue)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	// Test with nil client - this should panic or return an error
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling GithubRepoID with nil client")
		}
	}()

	GithubRepoID(context.Background(), nil)
}

func TestErrGitHubRepoNotFound(t *testing.T) {
	// Test that the error variable is properly defined
	if ErrGitHubRepoNotFound == nil {
		t.Error("ErrGitHubRepoNotFound should not be nil")
	}

	expectedMessage := "github repo not found"
	if ErrGitHubRepoNotFound.Error() != expectedMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, ErrGitHubRepoNotFound.Error())
	}
}
