package scan

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/catalog"
	"github.com/machinebox/graphql"
)

func (s *Scan) AssociateGithubPullRequest(ctx context.Context) (int, error) {
	prNum := pullRequestNumber()
	if prNum < 1 {
		return 0, nil
	}
	ghRepoID, err := catalog.GithubRepoID(ctx, s.client)
	if err != nil {
		return 0, err
	}

	req := graphql.NewRequest(`mutation AssociateGitHubPullRequest($scan_id: uuid!, $github_repo_id: uuid!, $github_pr_number: int_32!) {
  insert_vulnerability_reports_by_github_pull_request(
    objects: {scan_id: $scan_id, github_repo_id: $github_repo_id, github_pull_request_number: $github_pr_number}
  ) {
    returning {
      id
    }
  }
}`)
	req.Var("scan_id", s.ID)
	req.Var("github_repo_id", ghRepoID)
	req.Var("github_pr_number", prNum)

	var response struct {
		InsertVulnerabilityReportsByGithubPullRequest struct {
			Returning []struct {
				ID string `json:"id"`
			} `json:"returning"`
		} `json:"insert_vulnerability_reports_by_github_pull_request"`
	}

	return prNum, s.client.ExecuteGQL(ctx, req, &response)
}

func pullRequestNumber() int {
	var (
		buildkitBranch  = os.Getenv("BUILDKITE_PULL_REQUEST")
		githubRef       = os.Getenv("GITHUB_REF")
		branchRefPrefix = "refs/pull/"
	)
	var prNum string
	switch {
	case buildkitBranch != "":
		prNum = buildkitBranch
	case strings.HasPrefix(githubRef, branchRefPrefix):
		prNum = strings.TrimPrefix(githubRef, branchRefPrefix)
	}
	pr, _ := strconv.Atoi(prNum)
	return pr
}
