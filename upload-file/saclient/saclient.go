// Package saclient provides a client for interacting with the Security Agent API.
package saclient

import (
	"context"
	"net/http"
	"time"

	"github.com/machinebox/graphql"
)

type Client struct {
	securityAgentAPIEndpoint string
	securityAgentAPIKey      string

	gqlClient  *graphql.Client
	httpClient *http.Client
}

func NewClient(securityAgentAPIEndpoint, securityAgentAPIKey string) *Client {
	return &Client{
		securityAgentAPIEndpoint: securityAgentAPIEndpoint,
		securityAgentAPIKey:      securityAgentAPIKey,
		gqlClient:                graphql.NewClient(securityAgentAPIEndpoint),
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Allow up to 5 minutes for upload
		},
	}
}

func (c *Client) Do(ctx context.Context, req *graphql.Request, response interface{}) error {
	req.Header.Set("Authorization", c.securityAgentAPIKey)
	req.Header.Set("X-Hasura-Auth-Mode", "ci-auth")

	return c.gqlClient.Run(ctx, req, &response)
}
