package scan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
)

func TestScan_AssociateScanReport_Success(t *testing.T) {
	expectedReportPath := "s3://bucket/reports/scan-123.json"

	// Mock GraphQL server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"insert_vulnerability_reports_scan_reports": {
					"returning": [{
						"id": "report-assoc-id-789"
					}]
				}
			}
		}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	err := scan.AssociateScanReport(context.Background(), expectedReportPath)

	if err != nil {
		t.Errorf("AssociateScanReport failed: %v", err)
	}
}

func TestScan_AssociateScanReport_GraphQLError(t *testing.T) {
	// Mock GraphQL server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Database error"}]}`))
	}))
	defer server.Close()

	client := saclient.NewClient(server.URL, "test-api-key")
	scan := &Scan{
		ID:     "scan-id-123",
		Tags:   map[string]string{},
		client: client,
	}

	err := scan.AssociateScanReport(context.Background(), "s3://bucket/reports/scan-123.json")

	if err == nil {
		t.Error("Expected error for GraphQL error response")
	}
}
