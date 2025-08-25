package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	ServiceMetadataUploadPath = "metadata/services"
)

func ServiceMetadata(ctx context.Context, c *Client, serviceName, sourceCodePath, dockerfilePath string) error {
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
			SourceCodePath: sourceCodePath,
			DockerfilePath: dockerfilePath,
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
	serviceMetadataUploadPath := ServiceMetadataUploadPath + "/" + serviceName + ".json"
	err = c.UploadFile(ctx, metadataFile.Name(), serviceMetadataUploadPath)
	if err != nil {
		return fmt.Errorf("failed to upload metadata: %v", err)
	}
	log.Println("Service metadata upload completed successfully")

	return nil
}
