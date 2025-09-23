package saclient

import (
	"testing"
)

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected ContentType
	}{
		{
			name:     "JSON file",
			filePath: "data.json",
			expected: ContentTypeJSON,
		},
		{
			name:     "JSON file with path",
			filePath: "/path/to/file.json",
			expected: ContentTypeJSON,
		},
		{
			name:     "Text file",
			filePath: "readme.txt",
			expected: ContentTypeText,
		},
		{
			name:     "Text file with path",
			filePath: "/home/user/document.txt",
			expected: ContentTypeText,
		},
		{
			name:     "CSV file",
			filePath: "data.csv",
			expected: ContentTypeCSV,
		},
		{
			name:     "CSV file with path",
			filePath: "/exports/report.csv",
			expected: ContentTypeCSV,
		},
		{
			name:     "XML file",
			filePath: "config.xml",
			expected: ContentTypeXML,
		},
		{
			name:     "XML file with path",
			filePath: "/config/settings.xml",
			expected: ContentTypeXML,
		},
		{
			name:     "PDF file",
			filePath: "document.pdf",
			expected: ContentTypePDF,
		},
		{
			name:     "PDF file with path",
			filePath: "/documents/report.pdf",
			expected: ContentTypePDF,
		},
		{
			name:     "ZIP file",
			filePath: "archive.zip",
			expected: ContentTypeZIP,
		},
		{
			name:     "ZIP file with path",
			filePath: "/downloads/backup.zip",
			expected: ContentTypeZIP,
		},
		{
			name:     "TAR file",
			filePath: "backup.tar",
			expected: ContentTypeTAR,
		},
		{
			name:     "TAR file with path",
			filePath: "/backups/data.tar",
			expected: ContentTypeTAR,
		},
		{
			name:     "GZ file",
			filePath: "compressed.gz",
			expected: ContentTypeGZ,
		},
		{
			name:     "GZ file with path",
			filePath: "/tmp/archive.gz",
			expected: ContentTypeGZ,
		},
		{
			name:     "Unknown extension",
			filePath: "file.unknown",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "No extension",
			filePath: "filename",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "No extension with path",
			filePath: "/path/to/filename",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "Empty filename",
			filePath: "",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "Dot file without extension",
			filePath: ".hidden",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "Dot file with extension",
			filePath: ".config.json",
			expected: ContentTypeJSON,
		},
		{
			name:     "Multiple dots in filename",
			filePath: "file.backup.json",
			expected: ContentTypeJSON,
		},
		{
			name:     "Case sensitivity - uppercase extension",
			filePath: "file.JSON",
			expected: ContentTypeOctetStream, // Should not match as extensions are case-sensitive
		},
		{
			name:     "Case sensitivity - mixed case extension",
			filePath: "file.Json",
			expected: ContentTypeOctetStream, // Should not match as extensions are case-sensitive
		},
		{
			name:     "Binary file extension",
			filePath: "program.exe",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "Image file extension",
			filePath: "photo.jpg",
			expected: ContentTypeOctetStream,
		},
		{
			name:     "Complex path with known extension",
			filePath: "/very/long/path/with/many/directories/file.csv",
			expected: ContentTypeCSV,
		},
		{
			name:     "Windows-style path",
			filePath: "C:\\Users\\Documents\\data.xml",
			expected: ContentTypeXML,
		},
		{
			name:     "File with spaces in name",
			filePath: "my document.pdf",
			expected: ContentTypePDF,
		},
		{
			name:     "File with special characters",
			filePath: "file-name_with@special#chars.txt",
			expected: ContentTypeText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContentType(tt.filePath)
			if result != tt.expected {
				t.Errorf("getContentType(%q) = %v, want %v", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestContentTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant ContentType
		expected string
	}{
		{
			name:     "ContentTypeJSON",
			constant: ContentTypeJSON,
			expected: "application/json",
		},
		{
			name:     "ContentTypeText",
			constant: ContentTypeText,
			expected: "text/plain",
		},
		{
			name:     "ContentTypeCSV",
			constant: ContentTypeCSV,
			expected: "text/csv",
		},
		{
			name:     "ContentTypeXML",
			constant: ContentTypeXML,
			expected: "application/xml",
		},
		{
			name:     "ContentTypePDF",
			constant: ContentTypePDF,
			expected: "application/pdf",
		},
		{
			name:     "ContentTypeZIP",
			constant: ContentTypeZIP,
			expected: "application/zip",
		},
		{
			name:     "ContentTypeTAR",
			constant: ContentTypeTAR,
			expected: "application/x-tar",
		},
		{
			name:     "ContentTypeGZ",
			constant: ContentTypeGZ,
			expected: "application/gzip",
		},
		{
			name:     "ContentTypeOctetStream",
			constant: ContentTypeOctetStream,
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.constant)
			if result != tt.expected {
				t.Errorf("ContentType constant %s = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}
