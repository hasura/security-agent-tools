package main

import (
	"context"
	"log"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/catalog"
	"github.com/hasura/security-agent-tools/upload-file/input"
	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/hasura/security-agent-tools/upload-file/scan"
)

func main() {
	input, err := input.Parse()
	if err != nil {
		log.Fatalln(err)
		return
	}

	secAgentClient := saclient.NewClient(input.SecurityAgentAPIEndpoint, input.SecurityAgentAPIToken)

	err = secAgentClient.UploadFile(context.Background(), input.FilePath, input.Destination)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	log.Printf("Upload successful: %s -> %s\n", input.FilePath, input.Destination)

	sc, err := scan.New(context.Background(), secAgentClient, input.Tags)
	if err != nil {
		log.Fatalf("Failed to create scan: %v", err)
	}
	log.Printf("Scan created with ID: %s\n", sc.ID)

	err = sc.AssociateScanReport(context.Background(), input.Destination)
	if err != nil {
		log.Fatalf("Failed to store scan report path in metadata: %v", err)
	}

	imageName, err := sc.AssociateImageName(context.Background())
	switch {
	case err != nil:
		log.Fatalf("Failed to associate image name with scan: %v", err)
	case imageName != "":
		log.Printf("Associated image name: %s\n", imageName)
	}

	domain, err := sc.AssociateProductDomains(context.Background())
	switch {
	case err != nil:
		pd, _ := catalog.ProductDomains(context.Background(), secAgentClient)
		var pds strings.Builder
		for _, p := range pd {
			pds.WriteString("  - " + p + "\n")
		}
		log.Fatalf("Failed to associate product domain with scan: %v. Please check `product_domain` value is one of the following:\n%s", err, pds.String())
	case domain != "":
		log.Printf("Associated product domain(s): %s\n", domain)
	}

	serviceName, err := sc.AssociateServiceName(context.Background())
	switch {
	case err != nil:
		log.Fatalf("Failed to associate service name with scan: %v", err)
	case serviceName != "":
		log.Printf("Associated service name: %s\n", serviceName)
	}

	githubBranchName, err := sc.AssociateGithubBranchName(context.Background())
	switch {
	case err != nil:
		log.Fatalf("Failed to associate GitHub branch name with scan: %v", err)
	case githubBranchName != "":
		log.Printf("Associated GitHub branch name: %s\n", githubBranchName)
	}

	githubPRNumber, err := sc.AssociateGithubPullRequest(context.Background())
	switch {
	case err != nil:
		log.Fatalf("Failed to associate GitHub pull request with scan: %v", err)
	case githubPRNumber > 0:
		log.Printf("Associated GitHub pull request: %d\n", githubPRNumber)
	}

	productRelease, err := sc.AssociateProductRelease(context.Background())
	switch {
	case err != nil:
		log.Fatalf("Failed to associate product release with scan: %v", err)
	case productRelease != "":
		log.Printf("Associated product release: %s\n", productRelease)
	}
}
