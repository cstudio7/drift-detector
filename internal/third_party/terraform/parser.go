package terraform

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cstudio7/drift-detector/internal/third_party/logger"
)

// TFStateParser defines the interface for parsing Terraform state files.
type TFStateParser interface {
	ParseTFState(filePath string) (InstanceConfigSet, error)
}

// TFStateParserImpl is the implementation of TFStateParser.
type TFStateParserImpl struct {
	logger logger.Logger
}

// NewTFStateParser creates a new TFStateParserImpl.
func NewTFStateParser(logger logger.Logger) *TFStateParserImpl {
	return &TFStateParserImpl{
		logger: logger,
	}
}

// InstanceConfigSet holds aggregated attributes for drift detection.
type InstanceConfigSet struct {
	InstanceTypes       []string `json:"instance_types"`
	AMIs                []string `json:"amis"`
	AvailabilityZones   []string `json:"availability_zones"`
	KeyNames            []string `json:"key_names"`
	SecurityGroupIDs    []string `json:"security_groups"`
	SubnetIDs           []string `json:"subnet_ids"`
	IAMInstanceProfiles []string `json:"iam_instance_profiles"`
	TagNames            []string `json:"tag_names"`
	TagEnvironments     []string `json:"tag_environments"`
	EBSVolumeSizes      []int    `json:"ebs_volume_sizes"`
	EBSVolumeTypes      []string `json:"ebs_volume_types"`
}

// IsEmpty checks if an InstanceConfigSet is empty.
func (c InstanceConfigSet) IsEmpty() bool {
	return len(c.InstanceTypes) == 0 &&
		len(c.AMIs) == 0 &&
		len(c.AvailabilityZones) == 0 &&
		len(c.KeyNames) == 0 &&
		len(c.SecurityGroupIDs) == 0 &&
		len(c.SubnetIDs) == 0 &&
		len(c.IAMInstanceProfiles) == 0 &&
		len(c.TagNames) == 0 &&
		len(c.TagEnvironments) == 0 &&
		len(c.EBSVolumeSizes) == 0 &&
		len(c.EBSVolumeTypes) == 0
}

// TFState represents the structure of the simplified Terraform state file.
type TFState struct {
	Version          int    `json:"version"`
	TerraformVersion string `json:"terraform_version"`
	Serial           int    `json:"serial"`
	Lineage          string `json:"lineage"`
	Resources        struct {
		AWSInstance InstanceConfigSet `json:"aws_instance"`
	} `json:"resources"`
}

func (p *TFStateParserImpl) ParseTFState(filePath string) (InstanceConfigSet, error) {
	// Log the file being parsed
	p.logger.Info("Starting to parse file", "file_path", filePath)

	// Read the content of the JSON state file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		p.logger.Error("Failed to read JSON state file", "file_path", filePath, "error", err.Error())
		return InstanceConfigSet{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON content
	var state TFState
	if err := json.Unmarshal(fileContent, &state); err != nil {
		p.logger.Error("Failed to parse JSON state file", "file_path", filePath, "error", err.Error())
		return InstanceConfigSet{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Log parsed Terraform version
	p.logger.Info("Parsed Terraform state", "version", state.Version, "terraform_version", state.TerraformVersion)

	configSet := state.Resources.AWSInstance

	// Log aggregated attributes
	p.logger.Info("Parsed instance types", "instance_types", configSet.InstanceTypes)
	p.logger.Info("Parsed AMIs", "amis", configSet.AMIs)
	p.logger.Info("Parsed availability zones", "availability_zones", configSet.AvailabilityZones)
	p.logger.Info("Parsed key names", "key_names", configSet.KeyNames)
	p.logger.Info("Parsed security groups", "security_groups", configSet.SecurityGroupIDs)
	p.logger.Info("Parsed subnet IDs", "subnet_ids", configSet.SubnetIDs)
	p.logger.Info("Parsed IAM instance profiles", "iam_instance_profiles", configSet.IAMInstanceProfiles)
	p.logger.Info("Parsed tag names", "tag_names", configSet.TagNames)
	p.logger.Info("Parsed tag environments", "tag_environments", configSet.TagEnvironments)
	p.logger.Info("Parsed EBS volume sizes", "ebs_volume_sizes", configSet.EBSVolumeSizes)
	p.logger.Info("Parsed EBS volume types", "ebs_volume_types", configSet.EBSVolumeTypes)

	p.logger.Info("Returning parsed config set")
	return configSet, nil
}
