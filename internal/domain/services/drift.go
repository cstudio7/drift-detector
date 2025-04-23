package services

import (
	"reflect"
	"strings"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

type DriftService struct{}

func NewDriftService() *DriftService {
	return &DriftService{}
}

func (s *DriftService) DetectDrift(instanceID string, awsConfig, tfConfig entities.InstanceConfig, attributes []string) entities.DriftReport {
	report := entities.DriftReport{
		InstanceID: instanceID,
		Changes:    make(map[string]entities.Change),
	}

	for _, attr := range attributes {
		switch strings.ToLower(attr) {
		case "instance_type":
			if awsConfig.InstanceType != tfConfig.InstanceType {
				report.Changes["instance_type"] = entities.Change{
					Expected: tfConfig.InstanceType,
					Actual:   awsConfig.InstanceType,
				}
			}
		case "tags":
			if !reflect.DeepEqual(awsConfig.Tags, tfConfig.Tags) {
				report.Changes["tags"] = entities.Change{
					Expected: tfConfig.Tags,
					Actual:   awsConfig.Tags,
				}
			}
		case "security_groups":
			if !reflect.DeepEqual(awsConfig.SecurityGroupIDs, tfConfig.SecurityGroupIDs) {
				report.Changes["security_groups"] = entities.Change{
					Expected: tfConfig.SecurityGroupIDs,
					Actual:   awsConfig.SecurityGroupIDs,
				}
			}
		case "subnet_id":
			if awsConfig.SubnetID != tfConfig.SubnetID {
				report.Changes["subnet_id"] = entities.Change{
					Expected: tfConfig.SubnetID,
					Actual:   awsConfig.SubnetID,
				}
			}
		case "iam_instance_profile":
			if awsConfig.IAMInstanceProfile != tfConfig.IAMInstanceProfile {
				report.Changes["iam_instance_profile"] = entities.Change{
					Expected: tfConfig.IAMInstanceProfile,
					Actual:   awsConfig.IAMInstanceProfile,
				}
			}
		}
	}

	report.HasDrift = len(report.Changes) > 0
	return report
}
