package catalog

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

func ProductDomains(ctx context.Context, c *saclient.Client) ([]string, error) {
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

	err := c.ExecuteGQL(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	var productDomains []string
	for _, pd := range response.VulnerabilityReportsProductDomains {
		productDomains = append(productDomains, pd.Code)
	}

	return productDomains, nil
}
