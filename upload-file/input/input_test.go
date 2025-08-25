package input

import (
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
