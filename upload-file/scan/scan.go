package scan

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

type Scan struct {
	ID   string            `json:"id"`
	Tags map[string]string `json:"tags"`

	client *saclient.Client
}

func New(ctx context.Context, c *saclient.Client, tags map[string]string) (*Scan, error) {
	t := tags
	if t == nil {
		t = make(map[string]string)
	}

	req := graphql.NewRequest(`mutation CreateScan($tags: json) {
  insert_vulnerability_reports_scans(objects: {tags: $tags}) {
    returning {
      id
      tags
    }
  }
}`)
	req.Var("tags", t)

	var response struct {
		InsertVulnerabilityReportsScans struct {
			Returning []Scan `json:"returning"`
		} `json:"insert_vulnerability_reports_scans"`
	}

	err := c.ExecuteGQL(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	sc := &response.InsertVulnerabilityReportsScans.Returning[0]
	sc.client = c

	return sc, nil
}
