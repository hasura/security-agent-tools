package scan

import (
	"context"
	"errors"
	"strings"

	"github.com/machinebox/graphql"
)

func (s *Scan) associateProductDomain(ctx context.Context, productDomain string) (string, error) {
	if productDomain == "" {
		return "", nil
	}

	req := graphql.NewRequest(`mutation AssociateProductDomain($scan_id: uuid!, $product_domain: string!) {
  insert_vulnerability_reports_by_product_domains(
    objects: {scan_id: $scan_id, product_domain: $product_domain}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("product_domain", productDomain)

	var response struct {
		InsertVulnerabilityReportsByProductDomains struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_product_domains"`
	}

	return productDomain, s.client.ExecuteGQL(ctx, req, &response)
}

func (s *Scan) AssociateProductDomains(ctx context.Context) (string, error) {
	productDomain := s.Tags["product_domain"]
	if productDomain == "" {
		return "", nil
	}

	var errs []error
	for _, pd := range strings.Split(productDomain, ",") {
		_, err := s.associateProductDomain(ctx, strings.TrimSpace(pd))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return productDomain, errors.Join(errs...)
}
