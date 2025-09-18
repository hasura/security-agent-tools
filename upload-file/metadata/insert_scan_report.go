package metadata

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/upload"
	"github.com/machinebox/graphql"
)

func InsertScanReport(ctx context.Context, c *upload.Client, scanID, reportPath string) error {
	req := graphql.NewRequest(`mutation InsertScanReport($report_uri: string!) {
  insert_vulnerability_reports_scan_reports(objects: {report_uri: $report_uri}) {
    returning {
      id
    }
  }
}`)
	req.Var("report_uri", reportPath)

	var response struct {
		InsertVulnerabilityReportsScanReports struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_scan_reports"`
	}

	return c.Do(ctx, req, &response)
}
