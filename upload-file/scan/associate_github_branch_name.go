package scan

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateGithubBranchName(ctx context.Context) (string, error) {
	branchName := branchName()
	if branchName == "" {
		return "", nil
	}
	ghRepoID, err := githubRepoID(ctx, s.client)
	if err != nil {
		return "", err
	}

	req := graphql.NewRequest(`mutation AssociateGitHubBranchName($scan_id: uuid!, $github_repo_id: uuid!, $github_branch_name: string!) {
  insert_vulnerability_reports_by_github_branch_name(
    objects: {scan_id: $scan_id, github_repo_id: $github_repo_id, github_branch_name: $github_branch_name}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("github_repo_id", ghRepoID)
	req.Var("github_branch_name", branchName)

	var response struct {
		InsertVulnerabilityReportsByGithubBranchName struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_github_branch_name"`
	}

	return branchName, s.client.ExecuteGQL(ctx, req, &response)
}

func branchName() string {
	var (
		buildkitBranch  = os.Getenv("BUILDKITE_BRANCH")
		githubRef       = os.Getenv("GITHUB_REF")
		branchRefPrefix = "refs/heads/"
	)
	switch {
	case buildkitBranch != "":
		return buildkitBranch
	case strings.HasPrefix(githubRef, branchRefPrefix):
		return strings.TrimPrefix(githubRef, branchRefPrefix)
	default:
		return ""
	}
}

var errGitHubRepoNotFound = errors.New("github repo not found")

func githubRepoID(ctx context.Context, c *saclient.Client) (string, error) {
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
	case errGitHubRepoNotFound:
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
		return "", errGitHubRepoNotFound
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
