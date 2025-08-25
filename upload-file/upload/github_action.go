package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var (
	ErrNotInGitHubAction = fmt.Errorf("not in GitHub action")

	ghActionUploadPath  = "metadata/github-actions" + os.Getenv("GITHUB_REPOSITORY") + "/" + os.Getenv("GITHUB_REF") + "/" + os.Getenv("GITHUB_SHA") + ".json"
	githubActionEnvVars = []string{
		"GITHUB_JOB",
		"GITHUB_REF",
		"GITHUB_SHA",
		"GITHUB_REPOSITORY",
		"GITHUB_REPOSITORY_OWNER",
		"GITHUB_REPOSITORY_OWNER_ID",
		"GITHUB_RUN_ID",
		"GITHUB_RUN_NUMBER",
		"GITHUB_RETENTION_DAYS",
		"GITHUB_RUN_ATTEMPT",
		"GITHUB_ACTOR_ID",
		"GITHUB_ACTOR",
		"GITHUB_WORKFLOW",
		"GITHUB_HEAD_REF",
		"GITHUB_BASE_REF",
		"GITHUB_EVENT_NAME",
		"GITHUB_SERVER_URL",
		"GITHUB_API_URL",
		"GITHUB_GRAPHQL_URL",
		"GITHUB_REF_NAME",
		"GITHUB_REF_PROTECTED",
		"GITHUB_REF_TYPE",
		"GITHUB_WORKFLOW_REF",
		"GITHUB_WORKFLOW_SHA",
		"GITHUB_REPOSITORY_ID",
		"GITHUB_TRIGGERING_ACTOR",
		"GITHUB_WORKSPACE",
		"GITHUB_ACTION",
		"GITHUB_EVENT_PATH",
		"GITHUB_ACTION_REPOSITORY",
		"GITHUB_ACTION_REF",
		"GITHUB_PATH",
		"GITHUB_ENV",
		"GITHUB_STEP_SUMMARY",
		"GITHUB_STATE",
		"GITHUB_OUTPUT",
		"RUNNER_OS",
		"RUNNER_ARCH",
		"RUNNER_NAME",
		"RUNNER_ENVIRONMENT",
		"RUNNER_TOOL_CACHE",
		"RUNNER_TEMP",
		"RUNNER_WORKSPACE",
		"ACTIONS_RUNTIME_URL",
		"ACTIONS_RUNTIME_TOKEN",
		"ACTIONS_CACHE_URL",
		"ACTIONS_ID_TOKEN_REQUEST_URL",
		"ACTIONS_ID_TOKEN_REQUEST_TOKEN",
		"ACTIONS_RESULTS_URL",
		"GITHUB_ACTIONS",
		"CI",
	}
)

func GitHubActionMetadata(ctx context.Context, c *Client) error {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return ErrNotInGitHubAction
	}

	ghMetadata := make(map[string]string)
	for _, envVar := range githubActionEnvVars {
		ghMetadata[envVar] = os.Getenv(envVar)
	}
	metadataJSON, err := json.Marshal(ghMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	metadataFile, err := os.CreateTemp("", "metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create temp metadata file: %v", err)
	}
	defer os.Remove(metadataFile.Name())

	_, err = metadataFile.Write(metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to write metadata to temp file: %v", err)
	}

	log.Println("Uploading GitHub Action metadata")
	err = c.UploadFile(ctx, metadataFile.Name(), ghActionUploadPath)
	if err != nil {
		return fmt.Errorf("failed to upload metadata: %v", err)
	}
	log.Println("GitHub Action upload completed successfully")

	return nil
}
