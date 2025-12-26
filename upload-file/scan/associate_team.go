package scan

import (
	"context"

	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateTeam(ctx context.Context) (string, error) {
	team := s.Tags["team"]
	if team == "" {
		return "", nil
	}

	req := graphql.NewRequest(`mutation AssociateTeam($scan_id: uuid!, $team: string!) {
  insert_vulnerability_reports_by_team(
    objects: {scan_id: $scan_id, team: $team}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("team", team)

	var response struct {
		InsertVulnerabilityReportsByTeam struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_team"`
	}

	return team, s.client.ExecuteGQL(ctx, req, &response)
}
