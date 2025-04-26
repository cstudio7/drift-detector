package usecases

import (
	"fmt"
	"strings"

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

	d.logger.Info("Fetched AWS configs", "count", len(awsConfigs))

	// Parse Terraform state
	tfConfigs, err := d.tfParser.ParseTFState(tfStateFile)
	if err != nil {
		return fmt.Errorf("failed to parse Terraform state from file %s: %w", tfStateFile, err)
	}

	d.logger.Info("Parsed Terraform configs", "instance_types", tfConfigs.InstanceTypes)

	// Compare AWS and Terraform configs
	for _, awsConfig := range awsConfigs {
		d.logger.Info("AWS config", "instance_id", awsConfig.InstanceID, "instance_type", awsConfig.InstanceType)
		// Compare the configurations
		diff := compareConfigs(awsConfig, tfConfigs)
		if len(diff) > 0 {
			d.logger.Info("Drift detected", "instance_id", awsConfig.InstanceID, "changes", diff)
		}
	}

	return nil
}

// compareConfigs compares an AWS InstanceConfig against a Terraform InstanceConfigSet and returns a map of differences.
func compareConfigs(awsConfig entities.InstanceConfig, tfConfigs terraform.InstanceConfigSet) map[string]interface{} {
	diff := make(map[string]interface{})

	if !contains(tfConfigs.InstanceTypes, awsConfig.InstanceType) {
		diff["instance_type"] = map[string]interface{}{
			"aws": awsConfig.InstanceType,
			"tf":  strings.Join(tfConfigs.InstanceTypes, ", "),
		}
	}

	if !allIn(awsConfig.SecurityGroupIDs, tfConfigs.SecurityGroupIDs) {
		diff["security_group_ids"] = map[string]interface{}{
			"aws": awsConfig.SecurityGroupIDs,
			"tf":  strings.Join(tfConfigs.SecurityGroupIDs, ", "),
		}
	}

	if !contains(tfConfigs.SubnetIDs, awsConfig.SubnetID) {
		diff["subnet_id"] = map[string]interface{}{
			"aws": awsConfig.SubnetID,
			"tf":  strings.Join(tfConfigs.SubnetIDs, ", "),
		}
	}

	if !contains(tfConfigs.IAMInstanceProfiles, awsConfig.IAMInstanceProfile) {
		diff["iam_instance_profile"] = map[string]interface{}{
			"aws": awsConfig.IAMInstanceProfile,
			"tf":  strings.Join(tfConfigs.IAMInstanceProfiles, ", "),
		}
	}

	if awsConfig.Tags != nil {
		if name, ok := awsConfig.Tags["Name"]; ok && !contains(tfConfigs.TagNames, name) {
			diff["tag.Name"] = map[string]interface{}{
				"aws": name,
				"tf":  strings.Join(tfConfigs.TagNames, ", "),
			}
		}
		if env, ok := awsConfig.Tags["Environment"]; ok && !contains(tfConfigs.TagEnvironments, env) {
			diff["tag.Environment"] = map[string]interface{}{
				"aws": env,
				"tf":  strings.Join(tfConfigs.TagEnvironments, ", "),
			}
		}
	}

	// Add EBS validation if awsConfig.EBSBlockDevices exists
	for _, ebs := range awsConfig.EBSBlockDevices {
		if !containsInt(tfConfigs.EBSVolumeSizes, ebs.VolumeSize) {
			diff["ebs_volume_size"] = map[string]interface{}{
				"aws": ebs.VolumeSize,
				"tf":  joinInt(tfConfigs.EBSVolumeSizes),
			}
		}
		if !contains(tfConfigs.EBSVolumeTypes, ebs.VolumeType) {
			diff["ebs_volume_type"] = map[string]interface{}{
				"aws": ebs.VolumeType,
				"tf":  strings.Join(tfConfigs.EBSVolumeTypes, ", "),
			}
		}
	}

	return diff
}

// contains checks if an item is in a string slice.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsInt checks if an item is in an int slice.
func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// allIn checks if all items in actual are in allowed.
func allIn(actual, allowed []string) bool {
	for _, a := range actual {
		if !contains(allowed, a) {
			return false
		}
	}
	return true
}

// joinInt converts an int slice to a string.
func joinInt(slice []int) string {
	strs := make([]string, len(slice))
	for i, v := range slice {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, ", ")
}
