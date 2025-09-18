package metadata

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/upload"
	"github.com/machinebox/graphql"
)

func AssociateImageNameWithScan(ctx context.Context, c *upload.Client, scanID, imageName string) error {
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
	req.Var("scan_id", scanID)
	req.Var("image_name", imageName)

	var response struct {
		InsertVulnerabilityReportsByImageName struct {
			Returning []struct {
				ID     string `json:"id"`
				ScanID string `json:"scan_id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_image_name"`
	}

	return c.Do(ctx, req, &response)
}
