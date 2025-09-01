package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/input"
)

const (
	metadataJSONFileName      = "metadata.json"
	ServiceMetadataUploadPath = "metadata/services"
)

func ServiceMetadata(ctx context.Context, c *Client, in *input.Input) error {
	serviceName := in.Tags["service"]
	metadataUploadPath := in.MetadataUploadPath
	if serviceName == "" && metadataUploadPath == "" {
		log.Println("Skipping metadata upload. Possible fixes:")
		log.Println("")
		log.Println("Add `tags: service=my-service-name` to your workflow to upload service metadata")
		log.Println("---- or ----")
		log.Println("Add `metadata-upload-path: my/custom/path` to your workflow to upload metadata to a custom path")
		return nil
	}

	type Scm struct {
		RepoURL        string `json:"repo_url"`
		HTTPSCloneURL  string `json:"https_clone_url"`
		SSHCloneURL    string `json:"ssh_clone_url"`
		SourceCodePath string `json:"source_code_path"`
		DockerfilePath string `json:"docker_file_path"`
	}
	type Metadata struct {
		ServiceName string `json:"service_name"`
		Scm         Scm    `json:"scm"`
	}
	metadata := Metadata{
		ServiceName: serviceName,
		Scm: Scm{
			SourceCodePath: in.Tags["source_code_path"],
			DockerfilePath: in.Tags["docker_file_path"],
		},
	}

	if os.Getenv("GITHUB_REPOSITORY") != "" {
		metadata.Scm.RepoURL = "https://github.com/" + os.Getenv("GITHUB_REPOSITORY")
		metadata.Scm.HTTPSCloneURL = "https://github.com/" + os.Getenv("GITHUB_REPOSITORY") + ".git"
		metadata.Scm.SSHCloneURL = "git@github.com:" + os.Getenv("GITHUB_REPOSITORY") + ".git"
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	log.Println("Uploading Service metadata")
	uploadPath := ""
	switch {
	case serviceName != "":
		uploadPath = servicePath(serviceName, metadataJSONFileName)
	case metadataUploadPath != "":
		uploadPath = metadataPath(metadataUploadPath, metadataJSONFileName)
	}
	err = c.UploadViaReader(ctx, bytes.NewReader(metadataJSON), ContentTypeJSON, uploadPath)
	if err != nil {
		return fmt.Errorf("failed to upload metadata: %v", err)
	}
	log.Println("Service metadata upload completed successfully")

	switch err := uploadGitHubActionMetadata(context.Background(), c, in); err {
	case ErrNotInGitHubAction:
		log.Println("Skipping GitHub action metadata upload, as we are not in GitHub action")
	default:
		return err
	}

	switch err := uploadBuildkiteMetadata(context.Background(), c, in); err {
	case ErrNotInBuildkite:
		log.Println("Skipping Buildkite metadata upload, as we are not in Buildkite")
	default:
		return err
	}

	return nil
}

// servicePath returns the upload path for a given service and path.
// path is relative to the service directory.
func servicePath(serviceName, path string) string {
	return metadataPath("services", serviceName, path)
}

func metadataPath(parts ...string) string {
	return "metadata" + "/" + strings.Join(parts, "/")
}
