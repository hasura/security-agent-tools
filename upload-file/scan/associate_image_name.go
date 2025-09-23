package scan

import (
	"context"

	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateImageName(ctx context.Context) (string, error) {
	imageName := s.Tags["image_name"]
	if imageName == "" {
		return "", nil
	}

	req := graphql.NewRequest(`mutation AssociateImageName($scan_id: uuid, $image_name: string!) {
  insert_vulnerability_reports_by_image_name(
    objects: {scan_id: $scan_id, image_name: $image_name}
  ) {
    returning {
      id
      scan_id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("image_name", imageName)

	var response struct {
		InsertVulnerabilityReportsByImageName struct {
			Returning []struct {
				ID     string `json:"id"`
				ScanID string `json:"scan_id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_image_name"`
	}

	return imageName, s.client.Do(ctx, req, &response)
}
