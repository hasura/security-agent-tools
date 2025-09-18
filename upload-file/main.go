package main

import (
	"context"
	"log"

	"github.com/hasura/security-agent-tools/upload-file/input"
	"github.com/hasura/security-agent-tools/upload-file/metadata"
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
	log.Printf("Upload successful: %s -> %s\n", input.FilePath, input.Destination)

	scan, err := metadata.CreateScan(context.Background(), c, input.Tags)
	if err != nil {
		log.Fatalf("Failed to create scan: %v", err)
	}
	log.Printf("Scan created. ID: %s. Tags: %v\n", scan.ID, scan.Tags)

	err = metadata.InsertScanReport(context.Background(), c, scan.ID, input.Destination)
	if err != nil {
		log.Fatalf("Failed to store scan report path in metadata: %v", err)
	}

	err = upload.ServiceMetadata(context.Background(), c, input)
	if err != nil {
		log.Fatalf("Failed to upload metadata: %v", err)
	}
}
