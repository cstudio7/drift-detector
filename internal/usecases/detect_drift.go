package usecases

import (
	"context"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
)

// DetectDriftUseCase defines the use case for detecting drift.
type DetectDriftUseCase struct {
	driftService *entities.DriftService
	awsClient    aws.EC2Client
	tfParser     terraform.TFStateParser
	logger       logger.Logger
	instanceIDs  []string
	tfStatePath  string
	attributes   []string
}

// NewDetectDriftUseCase creates a new DetectDriftUseCase.
func NewDetectDriftUseCase(
	driftService *entities.DriftService,
	awsClient aws.EC2Client,
	tfParser terraform.TFStateParser,
	logger logger.Logger,
	instanceIDs []string,
	tfStatePath string,
	attributes []string,
) *DetectDriftUseCase {
	return &DetectDriftUseCase{
		driftService: driftService,
		awsClient:    awsClient,
		tfParser:     tfParser,
		logger:       logger,
		instanceIDs:  instanceIDs,
		tfStatePath:  tfStatePath,
		attributes:   attributes,
	}
}

// Execute runs the drift detection use case.
func (uc *DetectDriftUseCase) Execute(ctx context.Context) error {
	// Parse Terraform state
	tfConfigs, err := uc.tfParser.ParseTFState(uc.tfStatePath)
	if err != nil {
		uc.logger.Error("Failed to parse Terraform state", "error", err)
		return err
	}

	// Process each instance ID
	for _, instanceID := range uc.instanceIDs {
		// Get AWS configuration
		awsInstance, err := uc.awsClient.GetInstance(ctx, instanceID)
		if err != nil {
			if err == entities.ErrInstanceNotFound {
				uc.logger.Warn("Instance not found in AWS", "instance_id", instanceID)
				continue
			}
			uc.logger.Error("Failed to get instance from AWS", "instance_id", instanceID, "error", err)
			return err
		}

		awsConfig := uc.awsClient.ToInstanceConfig(awsInstance)

		// Assume one instance in Terraform state for simplicity (as per original main.go)
		if len(tfConfigs) == 0 {
			uc.logger.Warn("No instances found in Terraform state", "instance_id", instanceID)
			continue
		}
		tfConfig := tfConfigs[0] // Use the first instance

		// Detect drift
		report := uc.driftService.DetectDrift(instanceID, awsConfig, tfConfig, uc.attributes)
		if report.HasDrift {
			uc.logger.Info("Drift detected", "instance_id", instanceID, "changes", report.Changes)
		} else {
			uc.logger.Info("No drift detected", "instance_id", instanceID)
		}
	}

	return nil
}
