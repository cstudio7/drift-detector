package services

import (
	"reflect"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

// DriftService defines the interface for drift detection logic.
type DriftService interface {
	DetectDrift(instanceID string, awsConfig, tfConfig entities.InstanceConfig, attributes []string) entities.DriftReport
}

// DriftServiceImpl is the implementation of DriftService.
type DriftServiceImpl struct{}

// NewDriftService creates a new DriftServiceImpl.
func NewDriftService() DriftService {
	return &DriftServiceImpl{}
}

// DetectDrift compares AWS and Terraform configurations and returns a drift report.
func (s *DriftServiceImpl) DetectDrift(instanceID string, awsConfig, tfConfig entities.InstanceConfig, attributes []string) entities.DriftReport {
	report := entities.DriftReport{
		InstanceID: instanceID,
		HasDrift:   false,
		Changes:    make(map[string]entities.Change),
	}

	for _, attr := range attributes {
		switch attr {
		case "instance_type":
			if awsConfig.InstanceType != tfConfig.InstanceType {
				report.HasDrift = true
				report.Changes["instance_type"] = entities.Change{
					Expected: tfConfig.InstanceType,
					Actual:   awsConfig.InstanceType,
				}
			}
		case "tags":
			if !reflect.DeepEqual(awsConfig.Tags, tfConfig.Tags) {
				report.HasDrift = true
				report.Changes["tags"] = entities.Change{
					Expected: tfConfig.Tags,
					Actual:   awsConfig.Tags,
				}
			}
		case "security_groups":
			if !reflect.DeepEqual(awsConfig.SecurityGroupIDs, tfConfig.SecurityGroupIDs) {
				report.HasDrift = true
				report.Changes["security_groups"] = entities.Change{
					Expected: tfConfig.SecurityGroupIDs,
					Actual:   awsConfig.SecurityGroupIDs,
				}
			}
		case "subnet_id":
			if awsConfig.SubnetID != tfConfig.SubnetID {
				report.HasDrift = true
				report.Changes["subnet_id"] = entities.Change{
					Expected: tfConfig.SubnetID,
					Actual:   awsConfig.SubnetID,
				}
			}
		case "iam_instance_profile":
			if awsConfig.IAMInstanceProfile != tfConfig.IAMInstanceProfile {
				report.HasDrift = true
				report.Changes["iam_instance_profile"] = entities.Change{
					Expected: tfConfig.IAMInstanceProfile,
					Actual:   awsConfig.IAMInstanceProfile,
				}
			}
		}
	}

	return report
}
