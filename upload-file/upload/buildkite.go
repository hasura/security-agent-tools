package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hasura/security-agent-tools/upload-file/input"
)

var (
	ErrNotInBuildkite = errors.New("not in Buildkite")

	buildkiteEnvVars = []string{
		// Build Information
		"BUILDKITE_BUILD_ID",
		"BUILDKITE_BUILD_NUMBER",
		"BUILDKITE_BUILD_URL",
		"BUILDKITE_BUILD_CREATOR",
		"BUILDKITE_MESSAGE",
		"BUILDKITE_PULL_REQUEST",
		"BUILDKITE_PULL_REQUEST_BASE_BRANCH",
		"BUILDKITE_REBUILT_FROM_BUILD_ID",

		// Pipeline and Agent Information
		"BUILDKITE_PIPELINE_ID",
		"BUILDKITE_PIPELINE_SLUG",
		"BUILDKITE_PIPELINE_NAME",
		"BUILDKITE_ORGANIZATION_SLUG",
		"BUILDKITE_AGENT_ID",
		"BUILDKITE_AGENT_NAME",

		// Job Information
		"BUILDKITE_JOB_ID",
		"BUILDKITE_COMMAND",
		"BUILDKITE_COMMAND_EXIT_STATUS",
		"BUILDKITE_JOB_URL",
		"BUILDKITE_STEP_KEY",

		// Git and Repository Information
		"BUILDKITE_REPO",
		"BUILDKITE_COMMIT",
		"BUILDKITE_BRANCH",
		"BUILDKITE_TAG",
		"BUILDKITE_CLEAN_CHECKOUT",

		// Other Variables
		"BUILDKITE_BUILD_PATH",
		"BUILDKITE_ARTIFACT_UPLOAD_DESTINATION",
		"BUILDKITE_PLUGINS_PATH",
	}
)

func uploadBuildkiteMetadata(ctx context.Context, c *Client, in *input.Input) error {
	if os.Getenv("BUILDKITE") != "true" {
		return nil
	}

	type Metadata struct {
		ScanReportPath string            `json:"scan_report_path"`
		Env            map[string]string `json:"env"`
		Tags           map[string]string `json:"tags"`
	}

	metadata := Metadata{
		ScanReportPath: in.Destination,
		Env:            make(map[string]string),
		Tags:           in.Tags,
	}

	for _, envVar := range buildkiteEnvVars {
		metadata.Env[envVar] = os.Getenv(envVar)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}
	metadataFile, err := os.CreateTemp("", "buildkite-metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create temp metadata file: %v", err)
	}
	defer os.Remove(metadataFile.Name())
	_, err = metadataFile.Write(metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to write metadata to temp file: %v", err)
	}

	buildkiteBranch := os.Getenv("BUILDKITE_BRANCH")
	buildkiteTag := os.Getenv("BUILDKITE_TAG")
	buildkiteCommit := os.Getenv("BUILDKITE_COMMIT")
	buildkitePullRequest := os.Getenv("BUILDKITE_PULL_REQUEST")
	uploadPath := ""
	switch {
	case buildkiteBranch != "":
		uploadPath = fmt.Sprintf("branches/%s/%s.json", buildkiteBranch, buildkiteCommit)
	case buildkiteTag != "":
		uploadPath = fmt.Sprintf("tags/%s/%s.json", buildkiteTag, buildkiteCommit)
	case buildkitePullRequest != "false":
		uploadPath = fmt.Sprintf("pull-requests/%s/%s.json", buildkitePullRequest, buildkiteCommit)
	default:
		return errors.New("failed to determine upload path. Please set at least one of BUILDKITE_BRANCH, BUILDKITE_TAG, BUILDKITE_PULL_REQUEST env vars")
	}
	serviceName := in.Tags["service"]
	buildkitePipelineSlug := os.Getenv("BUILDKITE_PIPELINE_SLUG")
	uploadPath = servicePath(serviceName, fmt.Sprintf("buildkite/%s/%s", buildkitePipelineSlug, uploadPath))

	log.Println("Uploading Buildkite metadata")
	err = c.UploadFile(ctx, metadataFile.Name(), uploadPath)
	if err != nil {
		return fmt.Errorf("failed to upload metadata: %v", err)
	}
	log.Println("Buildkite metadata upload completed successfully")

	return nil
}
