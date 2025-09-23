package saclient

import "path/filepath"

type ContentType string

const (
	ContentTypeJSON        ContentType = "application/json"
	ContentTypeText        ContentType = "text/plain"
	ContentTypeCSV         ContentType = "text/csv"
	ContentTypeXML         ContentType = "application/xml"
	ContentTypePDF         ContentType = "application/pdf"
	ContentTypeZIP         ContentType = "application/zip"
	ContentTypeTAR         ContentType = "application/x-tar"
	ContentTypeGZ          ContentType = "application/gzip"
	ContentTypeOctetStream ContentType = "application/octet-stream"
)

func getContentType(filePath string) ContentType {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".json":
		return ContentTypeJSON
	case ".txt":
		return ContentTypeText
	case ".csv":
		return ContentTypeCSV
	case ".xml":
		return ContentTypeXML
	case ".pdf":
		return ContentTypePDF
	case ".zip":
		return ContentTypeZIP
	case ".tar":
		return ContentTypeTAR
	case ".gz":
		return ContentTypeGZ
	default:
		return ContentTypeOctetStream
	}
}
