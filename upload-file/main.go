package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Get inputs from environment variables (GitHub Actions way)
	filePath := os.Getenv("INPUT_FILE_PATH")
	destination := os.Getenv("INPUT_DESTINATION")

	if filePath == "" {
		log.Fatal("file-path input is required")
	}
	if destination == "" {
		log.Fatal("destination input is required")
	}

	// Perform the upload
	err := uploadFile(filePath, destination)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	fmt.Printf("Upload completed successfully\n")
}

func uploadFile(filePath, destination string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get file info for logging
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// TODO: implement uploading logic here
	log.Printf("Would upload file: %s", filePath)
	log.Printf("File size: %d bytes", fileInfo.Size())
	log.Printf("Destination: %s", destination)
	log.Printf("Filename: %s", filepath.Base(filePath))

	return nil
}
