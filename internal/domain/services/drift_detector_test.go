package services

import (
	"reflect"
	"testing"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

func TestDriftService_DetectDrift(t *testing.T) {
	driftService := NewDriftService()

	tests := []struct {
		name       string
		awsConfig  entities.InstanceConfig
		tfConfig   entities.InstanceConfig
		attributes []string
		expected   entities.DriftReport
	}{
		{
			name: "DriftDetected",
			awsConfig: entities.InstanceConfig{
				InstanceType:       "t2.small",
				Tags:               map[string]string{"Name": "prod-instance"},
				SecurityGroupIDs:   []string{"sg-12345678"},
				SubnetID:           "subnet-12345678",
				IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/prod-profile",
			},
			tfConfig: entities.InstanceConfig{
				InstanceType:       "t2.micro",
				Tags:               map[string]string{"Name": "test-instance"},
				SecurityGroupIDs:   []string{"sg-12345678"},
				SubnetID:           "subnet-87654321",
				IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/test-profile",
			},
			attributes: []string{"instance_type", "tags", "security_groups", "subnet_id", "iam_instance_profile"},
			expected: entities.DriftReport{
				InstanceID: "i-1234567890abcdef0",
				HasDrift:   true,
				Changes: map[string]entities.Change{
					"instance_type":        {Expected: "t2.micro", Actual: "t2.small"},
					"tags":                 {Expected: map[string]string{"Name": "test-instance"}, Actual: map[string]string{"Name": "prod-instance"}},
					"subnet_id":            {Expected: "subnet-87654321", Actual: "subnet-12345678"},
					"iam_instance_profile": {Expected: "arn:aws:iam::123456789012:instance-profile/test-profile", Actual: "arn:aws:iam::123456789012:instance-profile/prod-profile"},
				},
			},
		},
		{
			name: "NoDrift",
			awsConfig: entities.InstanceConfig{
				InstanceType:       "t2.micro",
				Tags:               map[string]string{"Name": "test-instance"},
				SecurityGroupIDs:   []string{"sg-12345678"},
				SubnetID:           "subnet-12345678",
				IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/test-profile",
			},
			tfConfig: entities.InstanceConfig{
				InstanceType:       "t2.micro",
				Tags:               map[string]string{"Name": "test-instance"},
				SecurityGroupIDs:   []string{"sg-12345678"},
				SubnetID:           "subnet-12345678",
				IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/test-profile",
			},
			attributes: []string{"instance_type", "tags", "security_groups", "subnet_id", "iam_instance_profile"},
			expected: entities.DriftReport{
				InstanceID: "i-1234567890abcdef0",
				HasDrift:   false,
				Changes:    map[string]entities.Change{},
			},
		},
		{
			name: "SubsetAttributes",
			awsConfig: entities.InstanceConfig{
				InstanceType:       "t2.small",
				Tags:               map[string]string{"Name": "prod-instance"},
				SecurityGroupIDs:   []string{"sg-12345678"},
				SubnetID:           "subnet-12345678",
				IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/test-profile",
			},
			tfConfig: entities.InstanceConfig{
				InstanceType:       "t2.micro",
				Tags:               map[string]string{"Name": "test-instance"},
				SecurityGroupIDs:   []string{"sg-12345678"},
				SubnetID:           "subnet-12345678",
				IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/test-profile",
			},
			attributes: []string{"instance_type", "tags"},
			expected: entities.DriftReport{
				InstanceID: "i-1234567890abcdef0",
				HasDrift:   true,
				Changes: map[string]entities.Change{
					"instance_type": {Expected: "t2.micro", Actual: "t2.small"},
					"tags":          {Expected: map[string]string{"Name": "test-instance"}, Actual: map[string]string{"Name": "prod-instance"}},
				},
			},
		},
		{
			name: "EmptyConfigs",
			awsConfig: entities.InstanceConfig{
				InstanceType:     "",
				Tags:             nil,
				SecurityGroupIDs: nil,
			},
			tfConfig: entities.InstanceConfig{
				InstanceType:     "",
				Tags:             nil,
				SecurityGroupIDs: nil,
			},
			attributes: []string{"instance_type", "tags", "security_groups"},
			expected: entities.DriftReport{
				InstanceID: "i-1234567890abcdef0",
				HasDrift:   false,
				Changes:    map[string]entities.Change{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := driftService.DetectDrift("i-1234567890abcdef0", tt.awsConfig, tt.tfConfig, tt.attributes)

			if report.HasDrift != tt.expected.HasDrift {
				t.Errorf("Expected HasDrift %v, got %v", tt.expected.HasDrift, report.HasDrift)
			}
			if len(report.Changes) != len(tt.expected.Changes) {
				t.Errorf("Expected %d changes, got %d", len(tt.expected.Changes), len(report.Changes))
			}
			for key, expectedChange := range tt.expected.Changes {
				if change, ok := report.Changes[key]; !ok {
					t.Errorf("Expected change for %s not found", key)
				} else {
					if !reflect.DeepEqual(change.Expected, expectedChange.Expected) {
						t.Errorf("For %s, expected Expected value %v, got %v", key, expectedChange.Expected, change.Expected)
					}
					if !reflect.DeepEqual(change.Actual, expectedChange.Actual) {
						t.Errorf("For %s, expected Actual value %v, got %v", key, expectedChange.Actual, change.Actual)
					}
				}
			}
		})
	}
}
