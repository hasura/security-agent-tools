package main

import (
	"context"
	"log"
	"strings"

	"github.com/hasura/security-agent-tools/upload-file/input"
	"github.com/hasura/security-agent-tools/upload-file/metadata"
	"github.com/hasura/security-agent-tools/upload-file/saclient"
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

	scan, err := metadata.CreateScan(context.Background(), secAgentClient, input.Tags)
	if err != nil {
		log.Fatalf("Failed to create scan: %v", err)
	}
	log.Printf("Scan created with ID: %s\n", scan.ID)

	err = metadata.InsertScanReport(context.Background(), secAgentClient, scan.ID, input.Destination)
	if err != nil {
		log.Fatalf("Failed to store scan report path in metadata: %v", err)
	}

	imageName := input.Tags["image_name"]
	if imageName != "" {
		err = metadata.AssociateImageNameWithScan(context.Background(), secAgentClient, scan.ID, imageName)
		if err != nil {
			log.Fatalf("Failed to associate image name with scan: %v", err)
		}
		log.Printf("Associated image name: %s\n", imageName)
	}

	domain := input.Tags["product_domain"]
	if domain != "" {
		err = metadata.AssociateProductDomainWithScan(context.Background(), secAgentClient, scan.ID, domain)
		if err != nil {
			pd, _ := metadata.ProductDomains(context.Background(), secAgentClient)
			var pds strings.Builder
			for _, p := range pd {
				pds.WriteString("  - " + p + "\n")
			}
			log.Fatalf("Failed to associate product domain with scan: %v. Please check `product_domain` value is one of the following:\n%s", err, pds.String())
		}
		log.Printf("Associated product domain: %s\n", domain)
	}

	serviceName := input.Tags["service"]
	if serviceName != "" {
		err = metadata.AssociateServiceNameWithScan(context.Background(), secAgentClient, scan.ID, serviceName)
		if err != nil {
			log.Fatalf("Failed to associate service name with scan: %v", err)
		}
		log.Printf("Associated service name: %s\n", serviceName)
	}
}
