package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hasura/security-agent-tools/upload-file/input"
)

const (
	ServiceMetadataUploadPath = "metadata/services"
)

func ServiceMetadata(ctx context.Context, c *Client, in *input.Input) error {
	serviceName := in.Tags["service"]
	if serviceName == "" {
		log.Println("No service name provided, skipping service metadata upload")
		log.Println("Add `tags: service=my-service-name` to your workflow to upload service metadata")
		return nil
	}

	type Scm struct {
		RepoURL        string `json:"repo_url"`
		HTTPSCloneURL  string `json:"https_clone_url"`
		SSHCloneURL    string `json:"ssh_clone_url"`
		SourceCodePath string `json:"source_code_path"`
		DockerfilePath string `json:"dockerfile_path"`
	}
	type Metadata struct {
		ServiceName string `json:"service_name"`
		Scm         Scm    `json:"scm"`
	}
	metadata := Metadata{
		ServiceName: serviceName,
		Scm: Scm{
			SourceCodePath: in.Tags["source_code_path"],
			DockerfilePath: in.Tags["dockerfile_path"],
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

	metadataFile, err := os.CreateTemp("", "service-metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create temp metadata file: %v", err)
	}
	defer os.Remove(metadataFile.Name())

	_, err = metadataFile.Write(metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to write metadata to temp file: %v", err)
	}

	log.Println("Uploading Service metadata")
	err = c.UploadFile(ctx, metadataFile.Name(), servicePath(serviceName, "metadata.json"))
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

	return nil
}

// servicePath returns the upload path for a given service and path.
// path is relative to the service directory.
func servicePath(serviceName, path string) string {
	return ServiceMetadataUploadPath + "/" + serviceName + "/" + path
}
