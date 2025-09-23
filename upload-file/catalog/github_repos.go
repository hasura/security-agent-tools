package catalog

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

var ErrGitHubRepoNotFound = errors.New("github repo not found")

func GithubRepoID(ctx context.Context, c *saclient.Client) (string, error) {
	githubRepo := os.Getenv("GITHUB_REPOSITORY")
	if githubRepo == "" {
		return "", nil
	}
	repoParts := strings.Split(githubRepo, "/")
	if len(repoParts) != 2 {
		return "", nil
	}
	orgName, repoName := repoParts[0], repoParts[1]

	ghRepoID, err := getGitHubRepo(ctx, c, orgName, repoName)
	switch err {
	case ErrGitHubRepoNotFound:
		return addGitHubRepo(ctx, c, orgName, repoName)
	default:
		return ghRepoID, err
	}
}

func getGitHubRepo(ctx context.Context, c *saclient.Client, orgName, repoName string) (string, error) {
	req := graphql.NewRequest(`query GetGitHubRepo($org_name: string!, $repo_name: string!) {
  service_catalog_github_repos_by_github_repos_org_name_repo_name_key(
    org_name: $org_name
    repo_name: $repo_name
  ) {
    id
  }
}`)
	req.Var("org_name", orgName)
	req.Var("repo_name", repoName)

	var response struct {
		ServiceCatalogGithubReposByGithubReposOrgNameRepoNameKey *struct {
			ID string `json:"id"`
		} `json:"service_catalog_github_repos_by_github_repos_org_name_repo_name_key"`
	}

	err := c.ExecuteGQL(ctx, req, &response)
	if err != nil {
		return "", err
	}

	if response.ServiceCatalogGithubReposByGithubReposOrgNameRepoNameKey == nil {
		return "", ErrGitHubRepoNotFound
	}

	return response.ServiceCatalogGithubReposByGithubReposOrgNameRepoNameKey.ID, nil
}

func addGitHubRepo(ctx context.Context, c *saclient.Client, orgName, repoName string) (string, error) {
	req := graphql.NewRequest(`mutation AddGitHubRepo($org_name: string!, $repo_name: string!) {
  insert_service_catalog_github_repos(objects: {org_name: $org_name, repo_name: $repo_name}) {
    returning {
      id
    }
  }
}`)
	req.Var("org_name", orgName)
	req.Var("repo_name", repoName)

	var response struct {
		InsertServiceCatalogGithubReposOne struct {
			ID string `json:"id"`
		} `json:"insert_service_catalog_github_repos"`
	}

	err := c.ExecuteGQL(ctx, req, &response)
	if err != nil {
		return "", err
	}

	return response.InsertServiceCatalogGithubReposOne.ID, nil
}
