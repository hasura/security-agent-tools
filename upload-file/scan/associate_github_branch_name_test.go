package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestBranchName_BuildkiteBranch(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment
	os.Setenv("BUILDKITE_BRANCH", "feature/test-branch")
	os.Setenv("GITHUB_REF", "refs/heads/main")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := branchName()
	expected := "feature/test-branch"

	if result != expected {
		t.Errorf("Expected branch name %s, got %s", expected, result)
	}
}

func TestBranchName_GithubRef(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment - no buildkite, only github
	os.Unsetenv("BUILDKITE_BRANCH")
	os.Setenv("GITHUB_REF", "refs/heads/develop")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := branchName()
	expected := "develop"

	if result != expected {
		t.Errorf("Expected branch name %s, got %s", expected, result)
	}
}

func TestBranchName_GithubRefInvalidPrefix(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithub := os.Getenv("GITHUB_REF")

	// Set test environment - invalid github ref format
	os.Unsetenv("BUILDKITE_BRANCH")
	os.Setenv("GITHUB_REF", "refs/tags/v1.0.0")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
	}()

	result := branchName()
	expected := ""

	if result != expected {
		t.Errorf("Expected empty branch name for invalid ref, got %s", result)
	}
}

func TestBranchName_NoEnvironmentVariables(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithub := os.Getenv("GITHUB_REF")

	// Clear both environment variables
	os.Unsetenv("BUILDKITE_BRANCH")
	os.Unsetenv("GITHUB_REF")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		}
	}()

	result := branchName()
	expected := ""

	if result != expected {
		t.Errorf("Expected empty branch name when no env vars set, got %s", result)
	}
}

func TestScan_AssociateGithubBranchName_Success(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithub := os.Getenv("GITHUB_REF")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_BRANCH", "feature/security-fix")
	os.Setenv("GITHUB_REPOSITORY", "hasura/security-agent-tools")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithub != "" {
			os.Setenv("GITHUB_REF", originalGithub)
		} else {
			os.Unsetenv("GITHUB_REF")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	callCount := 0
	expectedBranchName := "feature/security-fix"
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
			// Second call - associate branch name
			w.Write([]byte(`{
				"data": {
					"insert_vulnerability_reports_by_github_branch_name": {
						"returning": [{
							"id": "branch-assoc-id-456"
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

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubBranchName failed: %v", err)
	}

	if branchName != expectedBranchName {
		t.Errorf("Expected branch name %s, got %s", expectedBranchName, branchName)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 GraphQL calls, got %d", callCount)
	}
}

func TestScan_AssociateGithubBranchName_EmptyBranchName(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithub := os.Getenv("GITHUB_REF")

	// Clear environment variables
	os.Unsetenv("BUILDKITE_BRANCH")
	os.Unsetenv("GITHUB_REF")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
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

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubBranchName failed: %v", err)
	}

	if branchName != "" {
		t.Errorf("Expected empty branch name, got %s", branchName)
	}
}

func TestScan_AssociateGithubBranchName_GithubRepoIDError(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_BRANCH", "feature/test")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
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

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err == nil {
		t.Error("Expected error when GithubRepoID fails")
	}

	if branchName != "" {
		t.Errorf("Expected empty branch name when error occurs, got %s", branchName)
	}
}

func TestScan_AssociateGithubBranchName_GraphQLError(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_BRANCH", "feature/test")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
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
			// Second call - associate branch name (error)
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

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err == nil {
		t.Error("Expected error when GraphQL mutation fails")
	}

	if branchName != "feature/test" {
		t.Errorf("Expected branch name to be returned even on error, got %s", branchName)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 GraphQL calls, got %d", callCount)
	}
}

func TestScan_AssociateGithubBranchName_HTTPError(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_BRANCH", "feature/test")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
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
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err == nil {
		t.Error("Expected error for HTTP 500 response")
	}

	if branchName != "" {
		t.Errorf("Expected empty branch name when HTTP error occurs, got %s", branchName)
	}
}

func TestScan_AssociateGithubBranchName_EmptyGithubRepoID(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment - valid branch name but no GITHUB_REPOSITORY
	os.Setenv("BUILDKITE_BRANCH", "feature/awesome-feature")
	os.Unsetenv("GITHUB_REPOSITORY") // This will cause catalog.GithubRepoID to return empty string

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubBranchName failed: %v", err)
	}

	if branchName != "" {
		t.Errorf("Expected empty branch name when GitHub repo ID is empty, got %s", branchName)
	}
}

func TestScan_AssociateGithubBranchName_EmptyGithubRepoID_InvalidFormat(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment - valid branch name but invalid GITHUB_REPOSITORY format
	os.Setenv("BUILDKITE_BRANCH", "develop")
	os.Setenv("GITHUB_REPOSITORY", "invalid-repo-format") // This will cause catalog.GithubRepoID to return empty string

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
		}
		if originalGithubRepo != "" {
			os.Setenv("GITHUB_REPOSITORY", originalGithubRepo)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY")
		}
	}()

	client := saclient.NewClient("https://example.com/graphql", "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	branchName, err := scan.AssociateGithubBranchName(context.Background())

	if err != nil {
		t.Errorf("AssociateGithubBranchName failed: %v", err)
	}

	if branchName != "" {
		t.Errorf("Expected empty branch name when GitHub repo ID is empty due to invalid format, got %s", branchName)
	}
}

func TestScan_AssociateGithubBranchName_ContextCancellation(t *testing.T) {
	// Save original values
	originalBuildkite := os.Getenv("BUILDKITE_BRANCH")
	originalGithubRepo := os.Getenv("GITHUB_REPOSITORY")

	// Set test environment
	os.Setenv("BUILDKITE_BRANCH", "feature/test")
	os.Setenv("GITHUB_REPOSITORY", "hasura/test-repo")

	// Restore original values
	defer func() {
		if originalBuildkite != "" {
			os.Setenv("BUILDKITE_BRANCH", originalBuildkite)
		} else {
			os.Unsetenv("BUILDKITE_BRANCH")
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

	branchName, err := scan.AssociateGithubBranchName(ctx)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if branchName != "" {
		t.Errorf("Expected empty branch name when context is cancelled, got %s", branchName)
	}
}
