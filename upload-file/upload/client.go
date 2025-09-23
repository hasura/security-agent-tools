package upload

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

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

func presignedUploadURL(ctx context.Context, c *saclient.Client, destination string) (string, error) {
	req := graphql.NewRequest(`
		query UploadFile($name: String!) {
			storage_presigned_upload_url(name: $name) {
				url
				expired_at
			}
		}
	`)
	req.Var("name", destination)

	var response struct {
		StoragePresignedUploadURL struct {
			URL       string    `json:"url"`
			ExpiredAt time.Time `json:"expired_at"`
		} `json:"storage_presigned_upload_url"`
	}
	err := c.Do(ctx, req, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned upload URL: %w", err)
	}

	// TODO: send back expired_at and use that as a timeout for upload
	log.Printf("Presigned URL expires at: %s", response.StoragePresignedUploadURL.ExpiredAt)

	return response.StoragePresignedUploadURL.URL, nil
}

// UploadFile uploads the file to S3 (or other storage bucket configured in security agent)
// using the presigned URL.
//
// sourcePath is the path to the file to upload.
// destination is the destination path in the storage bucket.
func UploadFile(ctx context.Context, c *saclient.Client, sourcePath, destination string) error {
	log.Printf("Uploading file: %s", sourcePath)

	fileInfo, err := os.Stat(sourcePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", sourcePath)
	}
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	log.Printf("Filename: %s", filepath.Base(sourcePath))
	log.Printf("File size: %d bytes", fileInfo.Size())
	log.Printf("Destination: %s", destination)

	presignedURL, err := presignedUploadURL(ctx, c, destination)
	if err != nil {
		return fmt.Errorf("failed to get presigned upload URL: %w", err)
	}

	log.Printf("Got presigned URL, uploading file...")
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a buffer to store the file content
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	return rawUpload(ctx, c,
		presignedURL,
		getContentType(sourcePath), fileInfo.Size(),
		&buf,
	)
}

func rawUpload(ctx context.Context, c *saclient.Client, uploadURL string, contentType ContentType, contentSize int64, r io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, r)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.ContentLength = contentSize
	req.Header.Set("Content-Type", string(contentType))

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
