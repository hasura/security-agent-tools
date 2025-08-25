package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hasura/security-agent-tools/upload-file/input"
	"github.com/hasura/security-agent-tools/upload-file/upload"
)

func main() {
	input, err := input.Parse()
	if err != nil {
		log.Fatalln(err)
		return
	}

	c := upload.NewClient(input.SecurityAgentAPIEndpoint, input.SecurityAgentAPIToken)

	err = c.UploadFile(context.Background(), input.FilePath, input.Destination)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	switch err := upload.GitHubActionMetadata(context.Background(), c); err {
	case upload.ErrNotInGitHubAction:
		log.Println("Skipping GitHub action metadata upload, as we are not in GitHub action")
	default:
		log.Fatalf("Failed to upload metadata: %v", err)
	}

	if serviceName := input.Tags["service"]; serviceName != "" {
		err = upload.ServiceMetadata(context.Background(), c, serviceName)
		if err != nil {
			log.Fatalf("Failed to upload metadata: %v", err)
		}
	}

	fmt.Printf("Upload completed successfully\n")
}
