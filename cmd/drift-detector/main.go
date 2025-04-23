package main

import (
	"context"
	"log"

	"github.com/cstudio7/drift-detector/internal/domain/services"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
	"github.com/cstudio7/drift-detector/internal/usecases"
)

func main() {
	ctx := context.Background()

	// Initialize dependencies
	driftService := services.NewDriftService()

	awsClient, err := aws.NewEC2Client(ctx)
	if err != nil {
		log.Fatalf("Failed to create AWS client: %v", err)
	}

	tfParser := terraform.NewTFStateParser()
	logger := logger.NewStdLogger()

	// Configuration (replace with your instance IDs and Terraform state file path)
	instanceIDs := []string{"i-1234567890abcdef0"} // Replace with your EC2 instance ID
	tfStatePath := "terraform.tfstate"             // Replace with your Terraform state file path
	attributes := []string{"instance_type", "tags", "security_groups", "subnet_id", "iam_instance_profile"}

	// Create and execute the use case
	useCase := usecases.NewDetectDriftUseCase(
		driftService,
		awsClient,
		tfParser,
		logger,
		instanceIDs,
		tfStatePath,
		attributes,
	)

	if err := useCase.Execute(ctx); err != nil {
		log.Fatalf("Failed to execute drift detection: %v", err)
	}

	log.Println("Drift detection completed successfully")
}
