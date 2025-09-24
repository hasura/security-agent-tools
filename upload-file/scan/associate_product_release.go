package scan

import (
	"context"

	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateProductRelease(ctx context.Context) (string, error) {
	productRelease := s.Tags["product_release"]
	if productRelease == "" {
		return "", nil
	}

	req := graphql.NewRequest(`mutation AssociateProductRelease($scan_id: uuid!, $release_version: string!) {
  insert_vulnerability_reports_by_product_release(
    objects: {scan_id: $scan_id, release_version: $release_version}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("release_version", productRelease)

	var response struct {
		InsertVulnerabilityReportsByProductRelease struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_product_release"`
	}

	return productRelease, s.client.ExecuteGQL(ctx, req, &response)
}
