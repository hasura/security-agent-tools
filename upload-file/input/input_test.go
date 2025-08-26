package input

import (
	"os"
	"reflect"
	"testing"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "single tag",
			input:    "key1=value1",
			expected: map[string]string{"key1": "value1"},
		},
		{
			name:     "multiple tags",
			input:    "key1=value1\nkey2=value2\nkey3=value3",
			expected: map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
		},
		{
			name:     "tag with empty value",
			input:    "key1=\nkey2=value2",
			expected: map[string]string{"key1": "", "key2": "value2"},
		},
		{
			name:     "tag with empty key",
			input:    "=value1\nkey2=value2",
			expected: map[string]string{"": "value1", "key2": "value2"},
		},
		{
			name:     "tag without equals sign (invalid)",
			input:    "key1\nkey2=value2",
			expected: map[string]string{"key2": "value2"},
		},
		{
			name:     "tag with multiple equals signs",
			input:    "key1=value1=extra\nkey2=value2",
			expected: map[string]string{"key1": "value1=extra", "key2": "value2"},
		},
		{
			name:     "tags with leading/trailing spaces",
			input:    "  key1=value1  \n key2=value2 \n  key3=value3",
			expected: map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
		},
		{
			name:     "duplicate keys (last one wins)",
			input:    "key1=value1\nkey1=value2",
			expected: map[string]string{"key1": "value2"},
		},
		{
			name:     "special characters in values",
			input:    "url=https://example.com\npath=/home/user\nemail=test@example.com",
			expected: map[string]string{"url": "https://example.com", "path": "/home/user", "email": "test@example.com"},
		},
		{
			name:     "empty lines",
			input:    "\n\n\n",
			expected: map[string]string{},
		},
		{
			name:     "mixed valid and invalid tags with empty lines",
			input:    "valid=value\ninvalid\n\nanother=good\n=empty_key\nno_value=\n",
			expected: map[string]string{"valid": "value", "another": "good", "": "empty_key", "no_value": ""},
		},
		{
			name:     "values with commas",
			input:    "list=item1,item2,item3\nurl=https://example.com/path?param1=value1,param2=value2",
			expected: map[string]string{"list": "item1,item2,item3", "url": "https://example.com/path?param1=value1,param2=value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseTags(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSecurityAgentAPIEndpoint(t *testing.T) {
	// Create a temporary test file for the tests
	testFile := "test-file.json"
	testContent := `{"test": "data"}`

	// Clean up function
	cleanup := func() {
		os.Remove(testFile)
	}
	defer cleanup()

	// Create test file
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name             string
		envEndpoint      string
		envAPIKey        string
		envFilePath      string
		expectedEndpoint string
		expectError      bool
	}{
		{
			name:             "uses custom endpoint from environment variable",
			envEndpoint:      "https://custom-security-agent.example.com/graphql",
			envAPIKey:        "test-api-key",
			envFilePath:      testFile,
			expectedEndpoint: "https://custom-security-agent.example.com/graphql",
			expectError:      false,
		},
		{
			name:             "uses default endpoint when environment variable is empty",
			envEndpoint:      "",
			envAPIKey:        "test-api-key",
			envFilePath:      testFile,
			expectedEndpoint: "https://security-agent.ddn.pro.hasura.io/graphql",
			expectError:      false,
		},
		{
			name:             "uses default endpoint when environment variable is not set",
			envEndpoint:      "", // Will be unset in test
			envAPIKey:        "test-api-key",
			envFilePath:      testFile,
			expectedEndpoint: "https://security-agent.ddn.pro.hasura.io/graphql",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variables
			originalEndpoint := os.Getenv("INPUT_SECURITY_AGENT_API_ENDPOINT")
			originalAPIKey := os.Getenv("INPUT_SECURITY_AGENT_API_KEY")
			originalFilePath := os.Getenv("INPUT_FILE_PATH")

			// Clean up environment variables after test
			defer func() {
				if originalEndpoint != "" {
					os.Setenv("INPUT_SECURITY_AGENT_API_ENDPOINT", originalEndpoint)
				} else {
					os.Unsetenv("INPUT_SECURITY_AGENT_API_ENDPOINT")
				}
				if originalAPIKey != "" {
					os.Setenv("INPUT_SECURITY_AGENT_API_KEY", originalAPIKey)
				} else {
					os.Unsetenv("INPUT_SECURITY_AGENT_API_KEY")
				}
				if originalFilePath != "" {
					os.Setenv("INPUT_FILE_PATH", originalFilePath)
				} else {
					os.Unsetenv("INPUT_FILE_PATH")
				}
			}()

			// Set test environment variables
			if tt.envEndpoint != "" {
				os.Setenv("INPUT_SECURITY_AGENT_API_ENDPOINT", tt.envEndpoint)
			} else {
				os.Unsetenv("INPUT_SECURITY_AGENT_API_ENDPOINT")
			}
			os.Setenv("INPUT_SECURITY_AGENT_API_KEY", tt.envAPIKey)
			os.Setenv("INPUT_FILE_PATH", tt.envFilePath)

			// Call Parse function
			result, err := Parse()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check endpoint value if no error expected
			if !tt.expectError {
				if result.SecurityAgentAPIEndpoint != tt.expectedEndpoint {
					t.Errorf("SecurityAgentAPIEndpoint = %q, want %q",
						result.SecurityAgentAPIEndpoint, tt.expectedEndpoint)
				}
			}
		})
	}
}
