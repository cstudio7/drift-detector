package usecases

import (
	"context"
	"sync"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/domain/services"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
)

// DetectDriftUseCase handles the drift detection logic.
type DetectDriftUseCase struct {
	driftService services.DriftService
	awsClient    aws.EC2Client
	tfParser     terraform.TFStateParser
	logger       logger.Logger
	instanceIDs  []string
	tfStatePath  string
	attributes   []string
}

// NewDetectDriftUseCase creates a new DetectDriftUseCase.
func NewDetectDriftUseCase(
	driftService services.DriftService,
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

// Execute runs the drift detection use case with concurrency.
func (uc *DetectDriftUseCase) Execute(ctx context.Context) error {
	// Parse Terraform state (this is a one-time operation, not parallelized)
	tfConfigs, err := uc.tfParser.ParseTFState(uc.tfStatePath)
	if err != nil {
		uc.logger.Error("Failed to parse Terraform state", "error", err)
		return err
	}

	// Set up concurrency constructs
	const maxWorkers = 5 // Limit the number of concurrent AWS API calls
	instanceChan := make(chan string, len(uc.instanceIDs))
	errChan := make(chan error, len(uc.instanceIDs))
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for instanceID := range instanceChan {
				// Fetch AWS instance
				instance, err := uc.awsClient.GetInstance(ctx, instanceID)
				if err != nil {
					if err == entities.ErrInstanceNotFound {
						uc.logger.Warn("Instance not found in AWS", "instance_id", instanceID)
						continue
					}
					uc.logger.Error("Failed to get instance from AWS", "instance_id", instanceID, "error", err)
					errChan <- err
					return
				}

				// Convert AWS instance to config
				awsConfig := uc.awsClient.ToInstanceConfig(instance)

				// Find matching Terraform config
				var tfConfig *entities.InstanceConfig
				for _, cfg := range tfConfigs {
					if cfg.InstanceID == instanceID {
						tfConfig = &cfg
						break
					}
				}

				if tfConfig != nil {
					report := uc.driftService.DetectDrift(instanceID, awsConfig, *tfConfig, uc.attributes)
					if report.HasDrift {
						uc.logger.Info("Drift detected", "instance_id", instanceID, "changes", report.Changes)
					}
				} else {
					uc.logger.Warn("Instance not found in Terraform state", "instance_id", instanceID)
				}
			}
		}()
	}

	// Distribute instance IDs to workers
	for _, instanceID := range uc.instanceIDs {
		instanceChan <- instanceID
	}
	close(instanceChan)

	// Wait for all workers to finish
	wg.Wait()
	close(errChan)

	// Check for any errors from workers
	if err, ok := <-errChan; ok {
		return err
	}

	return nil
}
