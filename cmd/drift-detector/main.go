package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/cstudio7/drift-detector/internal/domain/services"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
	"github.com/cstudio7/drift-detector/internal/usecases"
)

func main() {
	instanceIDs := flag.String("instance-ids", "", "Comma-separated list of EC2 instance IDs")
	tfStatePath := flag.String("tf-state", "", "Path to Terraform state file")
	attributes := flag.String("attributes", "instance_type,tags,security_groups", "Comma-separated list of attributes to compare")
	flag.Parse()

	if *instanceIDs == "" || *tfStatePath == "" {
		log.Fatal("instance-ids and tf-state are required")
	}

	// Split comma-separated values
	instanceIDList := strings.Split(*instanceIDs, ",")
	attributeList := strings.Split(*attributes, ",")

	// Initialize dependencies
	ctx := context.Background()
	driftService := services.NewDriftService()
	awsClient, err := aws.NewEC2Client(ctx)
	if err != nil {
		log.Fatalf("Failed to create AWS EC2 client: %v", err)
	}
	tfParser := terraform.NewTFStateParser()
	logger := logger.NewStdLogger()

	// Create and execute use case
	useCase := usecases.NewDetectDriftUseCase(
		driftService,
		awsClient,
		tfParser,
		logger,
		instanceIDList,
		*tfStatePath,
		attributeList,
	)

	if err := useCase.Execute(ctx); err != nil {
		log.Fatalf("Failed to execute drift detection: %v", err)
	}
}
