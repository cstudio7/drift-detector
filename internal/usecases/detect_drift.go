package usecases

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
)

type DriftDetector struct {
	awsClient aws.AWSClient
	tfParser  terraform.TFStateParser
	logger    logger.Logger
}

func NewDriftDetector(awsClient aws.AWSClient, logger logger.Logger) *DriftDetector {
	return &DriftDetector{
		awsClient: awsClient,
		tfParser:  terraform.NewTFStateParser(logger),
		logger:    logger,
	}
}

// Getter methods for testing
func (d *DriftDetector) AWSClient() aws.AWSClient          { return d.awsClient }
func (d *DriftDetector) TFParser() terraform.TFStateParser { return d.tfParser }
func (d *DriftDetector) Logger() logger.Logger             { return d.logger }

func (d *DriftDetector) DetectDrift(tfStateFile string) error {
	awsConfigs, err := d.awsClient.FetchInstanceConfigs()
	if err != nil {
		return fmt.Errorf("%w: %v", entities.ErrFetchAWSConfigs, err)
	}
	if len(awsConfigs) == 0 {
		d.logger.Warn("No AWS configurations found")
		return entities.ErrEmptyConfigs
	}
	d.logger.Info("Fetched AWS configs", "count", len(awsConfigs))

	tfConfigs, err := d.tfParser.ParseTFState(tfStateFile)
	if err != nil {
		return fmt.Errorf("%w: %v", entities.ErrInvalidTerraformState, err)
	}
	if tfConfigs.IsEmpty() {
		d.logger.Warn("No Terraform configurations found")
		return entities.ErrEmptyConfigs
	}
	d.logger.Info("Parsed Terraform configs", "instance_types", tfConfigs.InstanceTypes)

	type driftResult struct {
		instanceID string
		diff       map[string]map[string]string
		err        error
	}
	results := make(chan driftResult, len(awsConfigs))

	var wg sync.WaitGroup
	for _, awsConfig := range awsConfigs {
		wg.Add(1)
		go func(config entities.InstanceConfig) {
			defer wg.Done()
			diff, err := compareConfigs(config, tfConfigs)
			results <- driftResult{
				instanceID: config.InstanceID,
				diff:       diff,
				err:        err,
			}
		}(awsConfig)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []error
	for result := range results {
		if result.err != nil {
			errs = append(errs, fmt.Errorf("instance %s: %w", result.instanceID, result.err))
			continue
		}
		if len(result.diff) > 0 {
			d.logger.Info(formatDrift(result.instanceID, result.diff))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %v", entities.ErrConfigComparison, errors.Join(errs...))
	}

	return nil
}

func compareConfigs(awsConfig entities.InstanceConfig, tfConfigs terraform.InstanceConfigSet) (map[string]map[string]string, error) {
	if awsConfig.InstanceID == "" {
		return nil, fmt.Errorf("%w: empty instance ID", entities.ErrConfigComparison)
	}

	diff := make(map[string]map[string]string)

	if !contains(tfConfigs.InstanceTypes, awsConfig.InstanceType) {
		diff["instance_type"] = map[string]string{"aws": awsConfig.InstanceType, "tf": strings.Join(tfConfigs.InstanceTypes, ", ")}
	}
	if !allIn(awsConfig.SecurityGroupIDs, tfConfigs.SecurityGroupIDs) {
		diff["security_group_ids"] = map[string]string{"aws": strings.Join(awsConfig.SecurityGroupIDs, ", "), "tf": strings.Join(tfConfigs.SecurityGroupIDs, ", ")}
	}
	if !contains(tfConfigs.SubnetIDs, awsConfig.SubnetID) {
		diff["subnet_id"] = map[string]string{"aws": awsConfig.SubnetID, "tf": strings.Join(tfConfigs.SubnetIDs, ", ")}
	}
	if !contains(tfConfigs.IAMInstanceProfiles, awsConfig.IAMInstanceProfile) {
		diff["iam_instance_profile"] = map[string]string{"aws": awsConfig.IAMInstanceProfile, "tf": strings.Join(tfConfigs.IAMInstanceProfiles, ", ")}
	}
	if awsConfig.Tags != nil {
		if name, ok := awsConfig.Tags["Name"]; ok && !contains(tfConfigs.TagNames, name) {
			diff["tag.Name"] = map[string]string{"aws": name, "tf": strings.Join(tfConfigs.TagNames, ", ")}
		}
		if env, ok := awsConfig.Tags["Environment"]; ok && !contains(tfConfigs.TagEnvironments, env) {
			diff["tag.Environment"] = map[string]string{"aws": env, "tf": strings.Join(tfConfigs.TagEnvironments, ", ")}
		}
	}
	for _, ebs := range awsConfig.EBSBlockDevices {
		if ebs.VolumeSize <= 0 {
			return nil, fmt.Errorf("%w: invalid EBS volume size %d", entities.ErrConfigComparison, ebs.VolumeSize)
		}
		if !containsInt(tfConfigs.EBSVolumeSizes, ebs.VolumeSize) {
			diff["ebs_volume_size"] = map[string]string{"aws": fmt.Sprintf("%d", ebs.VolumeSize), "tf": joinInt(tfConfigs.EBSVolumeSizes)}
		}
		if !contains(tfConfigs.EBSVolumeTypes, ebs.VolumeType) {
			diff["ebs_volume_type"] = map[string]string{"aws": ebs.VolumeType, "tf": strings.Join(tfConfigs.EBSVolumeTypes, ", ")}
		}
	}

	return diff, nil
}

func formatDrift(instanceID string, diff map[string]map[string]string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Drift detected for instance %s:\n", instanceID))
	for field, values := range diff {
		b.WriteString(fmt.Sprintf("  - %s: AWS=%s, Terraform=%s\n", field, values["aws"], values["tf"]))
	}
	return b.String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func allIn(actual, allowed []string) bool {
	for _, a := range actual {
		if !contains(allowed, a) {
			return false
		}
	}
	return true
}

func joinInt(slice []int) string {
	strs := make([]string, len(slice))
	for i, v := range slice {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, ", ")
}
