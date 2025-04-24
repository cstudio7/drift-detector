package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
)

// TFStateParser defines the interface for parsing Terraform state files.
type TFStateParser interface {
	ParseTFState(filePath string) ([]entities.InstanceConfig, error)
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

// ParseTFState parses a Terraform state file and returns a list of instance configurations.
func (p *TFStateParserImpl) ParseTFState(filePath string) ([]entities.InstanceConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		p.logger.Error("Failed to read Terraform state file", "file_path", filePath, "error", err)
		return nil, err
	}

	p.logger.Info("Read Terraform state file", "file_path", filePath, "size", len(data))

	var state struct {
		Resources []struct {
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Name      string `json:"name"`
			Provider  string `json:"provider"`
			Instances []struct {
				Attributes struct {
					ID                 string            `json:"id"`
					InstanceType       string            `json:"instance_type"`
					Tags               map[string]string `json:"tags"`
					SecurityGroups     []string          `json:"security_groups"`
					SubnetID           string            `json:"subnet_id"`
					IAMInstanceProfile string            `json:"iam_instance_profile"`
				} `json:"attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		p.logger.Error("Failed to unmarshal Terraform state", "error", err)
		return nil, err
	}

	p.logger.Info("Parsed Terraform state", "resource_count", len(state.Resources))

	var configs []entities.InstanceConfig
	for _, resource := range state.Resources {
		p.logger.Info("Processing resource", "mode", resource.Mode, "type", resource.Type, "name", resource.Name)
		if resource.Mode == "managed" && resource.Type == "aws_instance" {
			for _, instance := range resource.Instances {
				trimmedID := strings.TrimSpace(instance.Attributes.ID)
				p.logger.Info("Found aws_instance", "id", trimmedID)

				// Optional: Debug print to ensure no hidden characters
				fmt.Println("TF Instance ID:", fmt.Sprintf("%q", trimmedID))

				configs = append(configs, entities.InstanceConfig{
					InstanceID:         trimmedID,
					InstanceType:       strings.TrimSpace(instance.Attributes.InstanceType),
					Tags:               instance.Attributes.Tags,
					SecurityGroupIDs:   instance.Attributes.SecurityGroups,
					SubnetID:           strings.TrimSpace(instance.Attributes.SubnetID),
					IAMInstanceProfile: strings.TrimSpace(instance.Attributes.IAMInstanceProfile),
				})
			}
		}
	}

	p.logger.Info("Returning parsed configs", "count", len(configs))
	return configs, nil
}
