package scan

import (
	"context"

	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateScanReport(ctx context.Context, reportPath string) error {
	req := graphql.NewRequest(`mutation AssociateScanReport($scan_id: uuid, $report_uri: string!) {
  insert_vulnerability_reports_scan_reports(
    objects: {scan_id: $scan_id, report_uri: $report_uri}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("report_uri", reportPath)

	var response struct {
		InsertVulnerabilityReportsScanReports struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_scan_reports"`
	}

	return s.client.Do(ctx, req, &response)
}
