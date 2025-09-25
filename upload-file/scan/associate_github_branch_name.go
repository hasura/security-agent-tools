package scan

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/catalog"
	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateGithubBranchName(ctx context.Context) (string, error) {
	branchName := branchName()
	if branchName == "" {
		return "", nil
	}
	ghRepoID, err := catalog.GithubRepoID(ctx, s.client)
	if err != nil {
		return "", err
	}
	if ghRepoID == "" {
		log.Println("Unable to associate GitHub branch. Try setting GITHUB_REPOSITORY environment variable if not already set.")
		return "", nil
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
