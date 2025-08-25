package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ServiceMetadata(ctx context.Context, c *Client, serviceName string) error {
	metadata := map[string]string{
		"service_name": serviceName,
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
	serviceMetadataUploadPath := "metadata/service/" + serviceName + ".json"
	err = c.UploadFile(ctx, metadataFile.Name(), serviceMetadataUploadPath)
	if err != nil {
		return fmt.Errorf("failed to upload metadata: %v", err)
	}
	log.Println("Service metadata upload completed successfully")

	return nil
}
