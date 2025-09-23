package metadata

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/hasura/security-agent-tools/upload-file/upload"
	"github.com/machinebox/graphql"
)

func AssociateProductDomainWithScan(ctx context.Context, c *saclient.Client, scanID, productDomain string) error {
	req := graphql.NewRequest(`mutation AssociateProductDomain($scan_id: uuid!, $product_domain: string!) {
  insert_vulnerability_reports_by_product_domains(
    objects: {scan_id: $scan_id, product_domain: $product_domain}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", scanID)
	req.Var("product_domain", productDomain)

	var response struct {
		InsertVulnerabilityReportsByProductDomains struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_product_domains"`
	}

	return c.Do(ctx, req, &response)
}

func ProductDomains(ctx context.Context, c *upload.Client) ([]string, error) {
	req := graphql.NewRequest(`query GetProductDomains {
  vulnerability_reports_product_domains {
    code
  }
}`)

	var response struct {
		VulnerabilityReportsProductDomains []struct {
			Code string `json:"code"`
		} `json:"vulnerability_reports_product_domains"`
	}

	err := c.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	var productDomains []string
	for _, pd := range response.VulnerabilityReportsProductDomains {
		productDomains = append(productDomains, pd.Code)
	}

	return productDomains, nil
}
