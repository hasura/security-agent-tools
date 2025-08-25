package input

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Input struct {
	FilePath                 string
	Destination              string
	SecurityAgentAPIEndpoint string
	SecurityAgentAPIToken    string
	Tags                     map[string]string
}

var (
	ErrFilePath            = errors.New("file-path input is required")
	ErrSecurityAgentAPIKey = errors.New("security-agent-api-key input is required")
)

func Parse() (*Input, error) {
	input := &Input{}

	filePath := os.Getenv("INPUT_FILE_PATH")
	if filePath == "" {
		return nil, ErrFilePath
	}
	if filepath.Ext(filePath) != ".json" {
		log.Fatalf("file must be a JSON file, got: %s", filePath)
	}
	input.FilePath = filePath

	securityAgentAPIEndpoint := os.Getenv("INPUT_SECURITY_AGENT_API_ENDPOINT")
	if securityAgentAPIEndpoint == "" {
		input.SecurityAgentAPIEndpoint = "https://security-agent.ddn.pro.hasura.io/graphql"
	}
	input.SecurityAgentAPIEndpoint = securityAgentAPIEndpoint

	securityAgentAPIKey := os.Getenv("INPUT_SECURITY_AGENT_API_KEY")
	if securityAgentAPIKey == "" {
		return nil, ErrSecurityAgentAPIKey
	}
	input.SecurityAgentAPIToken = securityAgentAPIKey

	destination := os.Getenv("INPUT_DESTINATION")
	if destination == "" {
		// Calculate SHA256 of file contents
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()
		hash, err := calculateFileSHA256(file)
		if err != nil {
			return nil, err
		}
		destination = "uploads/" + hash + ".json"
	}
	input.Destination = destination

	tags := os.Getenv("INPUT_TAGS")
	if tags != "" {
		input.Tags = parseTags(tags)
	}

	return input, nil
}

func parseTags(tags string) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range strings.Split(tags, "\n") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		kv := strings.SplitN(tag, "=", 2)
		if len(kv) == 2 {
			tagMap[kv[0]] = kv[1]
		}
	}
	return tagMap
}

// calculateFileSHA256 calculates the SHA256 hash of a file's contents
func calculateFileSHA256(file *os.File) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %v", err)
	}

	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}
