package catalog

import (
	"context"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

func Teams(ctx context.Context, c *saclient.Client) ([]string, error) {
	req := graphql.NewRequest(`query GetTeams {
  team_catalog_teams {
    name
  }
}`)

	var response struct {
		Teams []struct {
			Name string `json:"name"`
		} `json:"team_catalog_teams"`
	}

	err := c.ExecuteGQL(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	var teams []string
	for _, team := range response.Teams {
		teams = append(teams, team.Name)
	}

	return teams, nil
}
