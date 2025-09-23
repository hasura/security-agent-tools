package scan

import (
	"context"

	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateServiceName(ctx context.Context) (string, error) {
	serviceName := s.Tags["service"]
	if serviceName == "" {
		return "", nil
	}

	req := graphql.NewRequest(`mutation AssociateServiceName($scan_id: uuid!, $service_name: string!) {
  insert_vulnerability_reports_by_service_name(
    objects: {scan_id: $scan_id, service_name: $service_name}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("service_name", serviceName)

	var response struct {
		InsertVulnerabilityReportsByServiceName struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_service_name"`
	}

	return serviceName, s.client.Do(ctx, req, &response)
}
