package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hasura/security-agent-tools/upload-file/input"
	"github.com/machinebox/graphql"
)

func main() {
	input, err := input.Parse()
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = uploadFile(input.FilePath, input.Destination, input.SecurityAgentAPIEndpoint, input.SecurityAgentAPIToken)
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

	type Metadata struct {
		OriginalFileName string            `json:"original_file_name"`
		FileContentPath  string            `json:"file_content_path"`
		GitHubMetadata   map[string]string `json:"github_metadata"`
		Tags             map[string]string `json:"tags"`
	}

	var metadata Metadata
	metadata.OriginalFileName = filepath.Base(filePath)
	metadata.FileContentPath = destination
	metadata.GitHubMetadata = make(map[string]string)
	metadata.Tags = make(map[string]string)

	githubActionEnvVars := []string{
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

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		metadataUploadPath := "metadata/" + os.Getenv("GITHUB_REPOSITORY") + "/" + os.Getenv("GITHUB_REF") + "/" + os.Getenv("GITHUB_SHA") + ".json"
		for _, envVar := range githubActionEnvVars {
			metadata.GitHubMetadata[envVar] = os.Getenv(envVar)
		}

		log.Println("Getting presigned URL for metadata upload", metadataUploadPath)
		presignedURL, err := getPresignedUploadURL(metadataUploadPath, securityAgentAPIEndpoint, securityAgentAPIKey)
		if err != nil {
			return fmt.Errorf("failed to get presigned upload URL: %v", err)
		}

		metadataJSON, err := json.Marshal(metadata)
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

		log.Println("Uploading metadata to S3")
		err = uploadFileToS3(metadataFile.Name(), presignedURL)
		if err != nil {
			return fmt.Errorf("failed to upload file to S3: %v", err)
		}

		log.Println("Metadata upload completed successfully")
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
	req.Header.Set("Authorization", securityAgentAPIKey)

	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the request with raw response capture
	var response PresignedUploadResponse

	err := client.Run(ctx, req, &response)
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
func uploadFileToS3(filePath string, presignedURL string) error {
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
