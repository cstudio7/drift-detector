package usecases

import (
	"fmt"
	"reflect"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
)

// DriftDetector is responsible for detecting drift between AWS and Terraform state.
type DriftDetector struct {
	awsClient aws.AWSClient
	tfParser  terraform.TFStateParser
	logger    logger.Logger
}

// NewDriftDetector creates a new DriftDetector.
func NewDriftDetector(awsClient aws.AWSClient, logger logger.Logger) *DriftDetector {
	return &DriftDetector{
		awsClient: awsClient,
		tfParser:  terraform.NewTFStateParser(logger),
		logger:    logger,
	}
}

// DetectDrift detects drift between AWS and Terraform state.
func (d *DriftDetector) DetectDrift(tfStateFile string) error {
	// Fetch AWS instance configurations
	awsConfigs, err := d.awsClient.FetchInstanceConfigs()
	if err != nil {
		return fmt.Errorf("failed to fetch AWS configs: %w", err)
	}

	d.logger.Info("Fetched AWS configs", "configs", awsConfigs)
	d.logger.Info("Terraform state file", "file", tfStateFile)

	// Parse Terraform state
	tfConfigs, err := d.tfParser.ParseTFState(tfStateFile)
	if err != nil {
		return fmt.Errorf("failed to parse Terraform state from file %s: %w", tfStateFile, err)
	}

	d.logger.Info("Parsed Terraform configs", "count", len(tfConfigs), tfConfigs)

	// Compare AWS and Terraform configs
	for _, awsConfig := range awsConfigs {
		d.logger.Info("AWS config", "instance_id", awsConfig.InstanceID, "instance_type", awsConfig.InstanceType)

		var matched bool
		for _, tfConfig := range tfConfigs {
			d.logger.Info("AWS config",
				"instance_id", awsConfig.InstanceID,
				"instance_type", awsConfig.InstanceType,
				"tags", awsConfig.Tags,
				"subnet_id", awsConfig.SubnetID,
				"security_groups", awsConfig.SecurityGroupIDs,
				"iam_instance_profile", awsConfig.IAMInstanceProfile,
			)

			if awsConfig.InstanceID == tfConfig.InstanceID {
				matched = true
				d.logger.Info("Comparing", "aws_instance_id", awsConfig.InstanceID, "tf_instance_id", tfConfig.InstanceID)

				// Compare the configurations
				if !reflect.DeepEqual(awsConfig, tfConfig) {
					diff := compareConfigs(awsConfig, tfConfig)
					d.logger.Info("Drift detected", "instance_id", awsConfig.InstanceID, "changes", diff)
				}
				break
			}
		}

		if !matched {
			d.logger.Info("Drift detected: Instance not found in Terraform state", "instance_id", awsConfig.InstanceID)
		}
	}

	// Check for instances in Terraform state but not in AWS
	for _, tfConfig := range tfConfigs {
		var matched bool
		for _, awsConfig := range awsConfigs {
			if tfConfig.InstanceID == awsConfig.InstanceID {
				matched = true
				break
			}
		}
		if !matched {
			d.logger.Info("Drift detected: Instance not found in AWS", "instance_id", tfConfig.InstanceID)
		}
	}

	return nil
}

// compareConfigs compares two InstanceConfig structs and returns a map of differences.
func compareConfigs(awsConfig, tfConfig entities.InstanceConfig) map[string]interface{} {
	diff := make(map[string]interface{})

	if awsConfig.InstanceType != tfConfig.InstanceType {
		fmt.Println("InstanceType mismatch:")
		fmt.Println("  AWS:", awsConfig.InstanceType)
		fmt.Println("  TF: ", tfConfig.InstanceType)
		diff["instance_type"] = map[string]string{"aws": awsConfig.InstanceType, "tf": tfConfig.InstanceType}
	}

	if !reflect.DeepEqual(awsConfig.Tags, tfConfig.Tags) {
		fmt.Println("Tags mismatch:")
		fmt.Println("  AWS:", awsConfig.Tags)
		fmt.Println("  TF: ", tfConfig.Tags)
		diff["tags"] = map[string]interface{}{"aws": awsConfig.Tags, "tf": tfConfig.Tags}
	}

	if !reflect.DeepEqual(awsConfig.SecurityGroupIDs, tfConfig.SecurityGroupIDs) {
		fmt.Println("SecurityGroupIDs mismatch:")
		fmt.Println("  AWS:", awsConfig.SecurityGroupIDs)
		fmt.Println("  TF: ", tfConfig.SecurityGroupIDs)
		diff["security_group_ids"] = map[string]interface{}{"aws": awsConfig.SecurityGroupIDs, "tf": tfConfig.SecurityGroupIDs}
	}

	if awsConfig.SubnetID != tfConfig.SubnetID {
		fmt.Println("SubnetID mismatch:")
		fmt.Println("  AWS:", awsConfig.SubnetID)
		fmt.Println("  TF: ", tfConfig.SubnetID)
		diff["subnet_id"] = map[string]string{"aws": awsConfig.SubnetID, "tf": tfConfig.SubnetID}
	}

	if awsConfig.IAMInstanceProfile != tfConfig.IAMInstanceProfile {
		fmt.Println("IAMInstanceProfile mismatch:")
		fmt.Println("  AWS:", awsConfig.IAMInstanceProfile)
		fmt.Println("  TF: ", tfConfig.IAMInstanceProfile)
		diff["iam_instance_profile"] = map[string]string{"aws": awsConfig.IAMInstanceProfile, "tf": tfConfig.IAMInstanceProfile}
	}

	return diff
}
