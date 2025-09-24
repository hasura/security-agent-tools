package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestPullRequestNumber_BuildkitePR(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment
	os.Setenv("BUILDKITE_PULL_REQUEST", "123")
	os.Setenv("GITHUB_REF", "refs/pull/456/merge")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := pullRequestNumber()
	expected := 123

	if result != expected {
		t.Errorf("Expected PR number %d, got %d", expected, result)
	}
}

func TestPullRequestNumber_GithubRefMerge(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment - no buildkite, only github ref with merge
	os.Unsetenv("BUILDKITE_PULL_REQUEST")
	os.Setenv("GITHUB_REF", "refs/pull/18/merge")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := pullRequestNumber()
	expected := 18

	if result != expected {
		t.Errorf("Expected PR number %d, got %d", expected, result)
	}
}

func TestPullRequestNumber_GithubRefHead(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment - github ref with head
	os.Unsetenv("BUILDKITE_PULL_REQUEST")
	os.Setenv("GITHUB_REF", "refs/pull/42/head")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := pullRequestNumber()
	expected := 42

	if result != expected {
		t.Errorf("Expected PR number %d, got %d", expected, result)
	}
}

func TestPullRequestNumber_GithubRefInvalidFormat(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment - invalid github ref format
	os.Unsetenv("BUILDKITE_PULL_REQUEST")
	os.Setenv("GITHUB_REF", "refs/heads/main")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := pullRequestNumber()
	expected := 0

	if result != expected {
		t.Errorf("Expected PR number %d for invalid ref, got %d", expected, result)
	}
}

func TestPullRequestNumber_NoEnvironmentVariables(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Clear both environment variables
	os.Unsetenv("BUILDKITE_PULL_REQUEST")
	os.Unsetenv("GITHUB_REF")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		}
	}()

	result := pullRequestNumber()
	expected := 0

	if result != expected {
		t.Errorf("Expected PR number %d when no env vars set, got %d", expected, result)
	}
}

func TestPullRequestNumber_InvalidBuildkiteValue(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment with invalid buildkite value
	os.Setenv("BUILDKITE_PULL_REQUEST", "false")
	os.Setenv("GITHUB_REF", "refs/pull/99/merge")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := pullRequestNumber()
	expected := 0 // "false" converts to 0

	if result != expected {
		t.Errorf("Expected PR number %d for invalid buildkite value, got %d", expected, result)
	}
}

func TestPullRequestNumber_ZeroPRNumber(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment with zero PR number
	os.Setenv("BUILDKITE_PULL_REQUEST", "0")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := pullRequestNumber()
	expected := 0

	if result != expected {
		t.Errorf("Expected PR number %d, got %d", expected, result)
	}
}

func TestScan_AssociateGithubPullRequest_Success(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_PULL_REQUEST", "123")
	os.Setenv("GITHUB_REPOSITORY", "hasura/security-agent-tools")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	callCount := 0
	expectedPRNumber := 123
	expectedRepoID := "test-repo-id-123"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if callCount == 1 {
			// First call - get GitHub repo
			w.Write([]byte(`{
				"data": {
					"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": {
						"id": "` + expectedRepoID + `"
					}
				}
			}`))
		} else {
			// Second call - associate PR
			w.Write([]byte(`{
				"data": {
					"insert_vulnerability_reports_by_github_pull_request": {
						"returning": [{
							"id": "pr-assoc-id-456"
						}]
					}
				}
			}`))
		}
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	prNumber, err := scan.AssociateGithubPullRequest(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubPullRequest failed: %v", err)
	}

	if prNumber != expectedPRNumber {
		t.Errorf("Expected PR number %d, got %d", expectedPRNumber, prNumber)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 GraphQL calls, got %d", callCount)
	}
}

func TestScan_AssociateGithubPullRequest_NoPRNumber(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithub := os.Getenv("GITHUB_REF")

	// Clear environment variables
	os.Unsetenv("BUILDKITE_PULL_REQUEST")
	os.Unsetenv("GITHUB_REF")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		}
	}()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	prNumber, err := scan.AssociateGithubPullRequest(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubPullRequest failed: %v", err)
	}

	if prNumber != 0 {
		t.Errorf("Expected PR number 0, got %d", prNumber)
	}
}

func TestScan_AssociateGithubPullRequest_ZeroPRNumber(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")

	// Set zero PR number
	os.Setenv("BUILDKITE_PULL_REQUEST", "0")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
	}()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	prNumber, err := scan.AssociateGithubPullRequest(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubPullRequest failed: %v", err)
	}

	if prNumber != 0 {
		t.Errorf("Expected PR number 0, got %d", prNumber)
	}
}

func TestScan_AssociateGithubPullRequest_GithubRepoIDError(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_PULL_REQUEST", "123")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database connection failed"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	prNumber, err := scan.AssociateGithubPullRequest(context.Background())

	if err == nil {
		t.Error("Expected error when GithubRepoID fails")
	}

	if prNumber != 0 {
		t.Errorf("Expected PR number 0 when error occurs, got %d", prNumber)
	}
}

func TestScan_AssociateGithubPullRequest_GraphQLError(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_PULL_REQUEST", "123")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	callCount := 0
	expectedRepoID := "test-repo-id-123"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if callCount == 1 {
			// First call - get GitHub repo (success)
			w.Write([]byte(`{
				"data": {
					"service_catalog_github_repos_by_github_repos_org_name_repo_name_key": {
						"id": "` + expectedRepoID + `"
					}
				}
			}`))
		} else {
			// Second call - associate PR (error)
			w.Write([]byte(`{"errors": [{"message": "Insert failed"}]}`))
		}
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	prNumber, err := scan.AssociateGithubPullRequest(context.Background())

	if err == nil {
		t.Error("Expected error when GraphQL mutation fails")
	}

	if prNumber != 123 {
		t.Errorf("Expected PR number to be returned even on error, got %d", prNumber)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 GraphQL calls, got %d", callCount)
	}
}

func TestScan_AssociateGithubPullRequest_ContextCancellation(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_PULL_REQUEST")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_PULL_REQUEST", "123")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_PULL_REQUEST", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_PULL_REQUEST")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	prNumber, err := scan.AssociateGithubPullRequest(ctx)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if prNumber != 0 {
		t.Errorf("Expected PR number 0 when context is cancelled, got %d", prNumber)
	}
}
