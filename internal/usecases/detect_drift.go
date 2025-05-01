package usecases

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/third_party/aws"
	"github.com/cstudio7/drift-detector/internal/third_party/logger"
	"github.com/cstudio7/drift-detector/internal/third_party/terraform"
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

// AWSClient Getter methods for testing
func (d *DriftDetector) AWSClient() aws.AWSClient {
	return d.awsClient
}

func (d *DriftDetector) TFParser() terraform.TFStateParser {
	return d.tfParser
}

func (d *DriftDetector) Logger() logger.Logger {
	return d.logger
}

// DetectDrift detects drift between AWS and Terraform state concurrently and logs structured output.
func (d *DriftDetector) DetectDrift(tfStateFile string) error {
	// Fetch AWS instance configurations
	awsConfigs, err := d.awsClient.FetchInstanceConfigs()
	if err != nil {
		return fmt.Errorf("%w: %v", entities.ErrFetchAWSConfigs, err)
	}

	if len(awsConfigs) == 0 {
		d.logger.Warn("No AWS configurations found")
		return entities.ErrEmptyConfigs
	}

	d.logger.Info("Fetched AWS configs", "count", len(awsConfigs))

	// Parse Terraform state
	tfConfigs, err := d.tfParser.ParseTFState(tfStateFile)
	if err != nil {
		return fmt.Errorf("%w: %v", entities.ErrInvalidTerraformState, err)
	}

	// Check for empty Terraform configs
	if tfConfigs.IsEmpty() {
		d.logger.Warn("No Terraform configurations found")
		return entities.ErrEmptyConfigs
	}

	d.logger.Info("Parsed Terraform configs", "instance_types", tfConfigs.InstanceTypes)

	// Channel to collect drift results
	type driftResult struct {
		instanceID string
		diff       map[string]interface{}
		err        error
	}
	results := make(chan driftResult, len(awsConfigs))

	// WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup

	// Process each AWS config concurrently
	for _, awsConfig := range awsConfigs {
		wg.Add(1)
		go func(config entities.InstanceConfig) {
			defer wg.Done()

			// Compare configurations
			diff, err := CompareConfigs(config, tfConfigs)
			results <- driftResult{
				instanceID: config.InstanceID,
				diff:       diff,
				err:        err,
			}
		}(awsConfig)
	}

	// Close results channel after all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and log results with structured output
	var comparisonErrors []error
	for result := range results {
		if result.err != nil {
			comparisonErrors = append(comparisonErrors, fmt.Errorf("instance %s: %w", result.instanceID, result.err))
			continue
		}
		if len(result.diff) > 0 {
			// Build structured drift message
			var driftDetails []string
			for field, details := range result.diff {
				d := details.(map[string]interface{})
				awsVal := fmt.Sprintf("%v", d["aws"])
				tfVal := fmt.Sprintf("%v", d["tf"])
				driftDetails = append(driftDetails, fmt.Sprintf("  - %s: AWS=%s, Terraform=%s", field, awsVal, tfVal))
			}
			// Log structured drift message
			d.logger.Info(fmt.Sprintf("Drift detected for instance %s:\n%s", result.instanceID, strings.Join(driftDetails, "\n")))
		}
	}

	// Return any comparison errors
	if len(comparisonErrors) > 0 {
		return fmt.Errorf("%w: %v", entities.ErrConfigComparison, fmt.Errorf("multiple errors: %v", comparisonErrors))
	}

	return nil
}

// CompareConfigs compares an AWS InstanceConfig against a Terraform InstanceConfigSet and returns a map of differences.
func CompareConfigs(awsConfig entities.InstanceConfig, tfConfigs terraform.InstanceConfigSet) (map[string]interface{}, error) {
	diff := make(map[string]interface{})

	// Validate AWS config
	if awsConfig.InstanceID == "" {
		return nil, fmt.Errorf("%w: empty instance ID", entities.ErrConfigComparison)
	}

	if !Contains(tfConfigs.InstanceTypes, awsConfig.InstanceType) {
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

	if !Contains(tfConfigs.SubnetIDs, awsConfig.SubnetID) {
		diff["subnet_id"] = map[string]interface{}{
			"aws": awsConfig.SubnetID,
			"tf":  strings.Join(tfConfigs.SubnetIDs, ", "),
		}
	}

	if !Contains(tfConfigs.IAMInstanceProfiles, awsConfig.IAMInstanceProfile) {
		diff["iam_instance_profile"] = map[string]interface{}{
			"aws": awsConfig.IAMInstanceProfile,
			"tf":  strings.Join(tfConfigs.IAMInstanceProfiles, ", "),
		}
	}

	if awsConfig.Tags != nil {
		if name, ok := awsConfig.Tags["Name"]; ok && !Contains(tfConfigs.TagNames, name) {
			diff["tag.Name"] = map[string]interface{}{
				"aws": name,
				"tf":  strings.Join(tfConfigs.TagNames, ", "),
			}
		}
		if env, ok := awsConfig.Tags["Environment"]; ok && !Contains(tfConfigs.TagEnvironments, env) {
			diff["tag.Environment"] = map[string]interface{}{
				"aws": env,
				"tf":  strings.Join(tfConfigs.TagEnvironments, ", "),
			}
		}
	}

	// Add EBS validation if awsConfig.EBSBlockDevices exists
	for _, ebs := range awsConfig.EBSBlockDevices {
		if ebs.VolumeSize <= 0 {
			return nil, fmt.Errorf("%w: invalid EBS volume size %d", entities.ErrConfigComparison, ebs.VolumeSize)
		}
		if !containsInt(tfConfigs.EBSVolumeSizes, ebs.VolumeSize) {
			diff["ebs_volume_size"] = map[string]interface{}{
				"aws": ebs.VolumeSize,
				"tf":  joinInt(tfConfigs.EBSVolumeSizes),
			}
		}
		if !Contains(tfConfigs.EBSVolumeTypes, ebs.VolumeType) {
			diff["ebs_volume_type"] = map[string]interface{}{
				"aws": ebs.VolumeType,
				"tf":  strings.Join(tfConfigs.EBSVolumeTypes, ", "),
			}
		}
	}

	return diff, nil
}

// Contains checks if an item is in a string slice.
func Contains(slice []string, item string) bool {
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
		if !Contains(allowed, a) {
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
