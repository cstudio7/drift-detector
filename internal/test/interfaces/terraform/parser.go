package terraform

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

// TFStateParser defines the interface for parsing Terraform state files.
type TFStateParser interface {
	ParseTFState(filePath string) ([]entities.InstanceConfig, error)
}

// TFStateParserImpl is the implementation of TFStateParser.
type TFStateParserImpl struct{}

// NewTFStateParser creates a new TFStateParserImpl.
func NewTFStateParser() *TFStateParserImpl {
	return &TFStateParserImpl{}
}

// TFState represents the structure of a Terraform state file.
type TFState struct {
	Resources []Resource `json:"resources"`
}

// Resource represents a Terraform resource.
type Resource struct {
	Type      string     `json:"type"`
	Instances []Instance `json:"instances"`
}

// Instance represents an instance within a Terraform resource.
type Instance struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// ParseTFState parses a Terraform state file and returns instance configurations.
func (p *TFStateParserImpl) ParseTFState(filePath string) ([]entities.InstanceConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Terraform state file: %w", err)
	}

	var state TFState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Terraform state: %w", err)
	}

	var configs []entities.InstanceConfig
	for _, resource := range state.Resources {
		if resource.Type != "aws_instance" {
			continue
		}
		for _, instance := range resource.Instances {
			config := entities.InstanceConfig{
				InstanceType:       instance.Attributes["instance_type"].(string),
				Tags:               make(map[string]string),
				SecurityGroupIDs:   make([]string, 0),
				SubnetID:           "",
				IAMInstanceProfile: "",
			}
			if tags, ok := instance.Attributes["tags"].(map[string]interface{}); ok {
				for k, v := range tags {
					config.Tags[k] = v.(string)
				}
			}
			if sg, ok := instance.Attributes["security_groups"].([]interface{}); ok {
				for _, id := range sg {
					config.SecurityGroupIDs = append(config.SecurityGroupIDs, id.(string))
				}
			}
			if subnet, ok := instance.Attributes["subnet_id"].(string); ok {
				config.SubnetID = subnet
			}
			if iam, ok := instance.Attributes["iam_instance_profile"].(string); ok {
				config.IAMInstanceProfile = iam
			}
			configs = append(configs, config)
		}
	}
	return configs, nil
}
