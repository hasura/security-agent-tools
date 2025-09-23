package metadata

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

func AssociateServiceNameWithScan(ctx context.Context, c *saclient.Client, scanID, serviceName string) error {
	req := graphql.NewRequest(`mutation AssociateServiceName($scan_id: uuid!, $service_name: string!) {
  insert_vulnerability_reports_by_service_name(
    objects: {scan_id: $scan_id, service_name: $service_name}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", scanID)
	req.Var("service_name", serviceName)

	var response struct {
		InsertVulnerabilityReportsByServiceName struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_service_name"`
	}

	return c.Do(ctx, req, &response)
}
