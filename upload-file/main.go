package main

import (
	"context"
	"log"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/input"
	"github.com/hasura/security-agent-tools/upload-file/saclient"
	"github.com/hasura/security-agent-tools/upload-file/scan"
	"github.com/hasura/security-agent-tools/upload-file/upload"
)

func main() {
	input, err := input.Parse()
	if err != nil {
		log.Fatalln(err)
		return
	}

	secAgentClient := saclient.NewClient(input.SecurityAgentAPIEndpoint, input.SecurityAgentAPIToken)

	err = upload.UploadFile(context.Background(), secAgentClient, input.FilePath, input.Destination)
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

	domain, err := sc.AssociateProductDomain(context.Background())
	switch {
	case err != nil:
		pd, _ := scan.ProductDomains(context.Background(), secAgentClient)
		var pds strings.Builder
		for _, p := range pd {
			pds.WriteString("  - " + p + "\n")
		}
		log.Fatalf("Failed to associate product domain with scan: %v. Please check `product_domain` value is one of the following:\n%s", err, pds.String())
	case domain != "":
		log.Printf("Associated product domain: %s\n", domain)
	}

	serviceName, err := sc.AssociateServiceName(context.Background())
	switch {
	case err != nil:
		log.Fatalf("Failed to associate service name with scan: %v", err)
	case serviceName != "":
		log.Printf("Associated service name: %s\n", serviceName)
	}
}
