package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/machinebox/graphql"
)

func main() {
	// Get inputs from environment variables (GitHub Actions way)
	filePath := os.Getenv("INPUT_FILE_PATH")
	destination := os.Getenv("INPUT_DESTINATION")
	securityAgentAPIEndpoint := os.Getenv("INPUT_SECURITY_AGENT_API_ENDPOINT")
	securityAgentAPIKey := os.Getenv("INPUT_SECURITY_AGENT_API_KEY")

	if filePath == "" {
		log.Fatal("file-path input is required")
	}
	if destination == "" {
		log.Fatal("destination input is required")
	}
	if securityAgentAPIKey == "" {
		log.Fatal("security-agent-api-key input is required")
	}

	// Perform the upload
	err := uploadFile(filePath, destination, securityAgentAPIEndpoint, securityAgentAPIKey)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	fmt.Printf("Upload completed successfully\n")
}

// GraphQL response structures
type PresignedUploadResponse struct {
	StoragePresignedUploadURL struct {
		URL       string    `json:"url"`
		ExpiredAt time.Time `json:"expired_at"`
	} `json:"storage_presigned_upload_url"`
}

func uploadFile(filePath, destination, securityAgentAPIEndpoint, securityAgentAPIKey string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Get file info for logging
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	log.Printf("Uploading file: %s", filePath)
	log.Printf("File size: %d bytes", fileInfo.Size())
	log.Printf("Destination: %s", destination)
	log.Printf("Filename: %s", filepath.Base(filePath))

	presignedURL, err := getPresignedUploadURL(destination, securityAgentAPIEndpoint, securityAgentAPIKey)
	if err != nil {
		return fmt.Errorf("failed to get presigned upload URL: %v", err)
	}

	log.Printf("Got presigned URL, uploading to S3...")
	err = uploadFileToS3(filePath, presignedURL)
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	log.Printf("File uploaded successfully to: %s", destination)
	return nil
}

// getPresignedUploadURL calls the GraphQL API to get a presigned upload URL
func getPresignedUploadURL(destination, securityAgentAPIEndpoint, securityAgentAPIKey string) (string, error) {
	log.Printf("Making GraphQL request to get presigned URL for destination: %s", destination)

	// Create GraphQL client
	client := graphql.NewClient(securityAgentAPIEndpoint)

	// Create the GraphQL request
	req := graphql.NewRequest(`
		query MyQuery($name: String!) {
			storage_presigned_upload_url(name: $name) {
				url
				expired_at
			}
		}
	`)
	req.Var("name", destination)
	req.Header.Set("Authorization", "Bearer "+securityAgentAPIKey)

	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the request with raw response capture
	var response PresignedUploadResponse
	var rawResponse interface{}

	// First, try to get the raw response to see what we're actually getting
	err := client.Run(ctx, req, &rawResponse)
	if err != nil {
		log.Printf("GraphQL request failed with error: %v", err)
		return "", fmt.Errorf("GraphQL request failed: %v", err)
	}

	// Log the raw response for debugging
	// rawJSON, _ := json.MarshalIndent(rawResponse, "", "  ")
	// log.Printf("Raw GraphQL response: %s", string(rawJSON))

	// Now try to parse into our expected structure
	err = client.Run(ctx, req, &response)
	if err != nil {
		log.Printf("Failed to parse GraphQL response into expected structure: %v", err)
		return "", fmt.Errorf("GraphQL request failed: %v", err)
	}

	// Check if we got a valid response
	if response.StoragePresignedUploadURL.URL == "" {
		log.Printf("Empty presigned URL received. Full response: %+v", response)
		return "", fmt.Errorf("empty presigned URL received from API")
	}

	log.Printf("Presigned URL expires at: %s", response.StoragePresignedUploadURL.ExpiredAt)
	return response.StoragePresignedUploadURL.URL, nil
}

// uploadFileToS3 uploads the file to S3 using the presigned URL
func uploadFileToS3(filePath, presignedURL string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get file info to determine content type and size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Create a buffer to store the file content
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Create HTTP request for PUT upload
	req, err := http.NewRequest("PUT", presignedURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set content length
	req.ContentLength = fileInfo.Size()

	// Determine content type based on file extension
	contentType := getContentType(filePath)
	req.Header.Set("Content-Type", contentType)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Allow up to 5 minutes for upload
	}

	// Execute the upload
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// getContentType determines the content type based on file extension
func getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	case ".csv":
		return "text/csv"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}
