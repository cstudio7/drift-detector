package terraform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
)

// getTestDataPath resolves the path to the testdata directory dynamically.
func getTestDataPath(filename string) (string, error) {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrInvalid
	}
	// Navigate to the testdata directory at the project root
	// internal/interfaces/terraform/terraform_test.go -> testdata/
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(callerFile)))) // Up four levels to project root
	testDataPath := filepath.Join(baseDir, "testdata", filename)
	return testDataPath, nil
}

func TestTFStateParserImpl_ParseTFState(t *testing.T) {
	// Sample Terraform state data
	validState := map[string]interface{}{
		"resources": []interface{}{
			map[string]interface{}{
				"mode":     "managed",
				"type":     "aws_instance",
				"name":     "example",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"id":                   "i-1234567890abcdef0",
							"instance_type":        "t2.micro",
							"tags":                 map[string]interface{}{"Name": "test-instance"},
							"security_groups":      []interface{}{"sg-12345678"},
							"subnet_id":            "subnet-12345678",
							"iam_instance_profile": "test-profile",
						},
					},
				},
			},
		},
	}

	emptyState := map[string]interface{}{
		"resources": []interface{}{},
	}

	// Create temporary files for testing
	validStateFile, err := getTestDataPath("valid-tfstate.json")
	if err != nil {
		t.Fatalf("Failed to resolve valid-tfstate.json path: %v", err)
	}
	validData, _ := json.Marshal(validState)
	if err := os.WriteFile(validStateFile, validData, 0644); err != nil {
		t.Fatalf("Failed to write valid-tfstate.json: %v", err)
	}
	defer os.Remove(validStateFile)

	emptyStateFile, err := getTestDataPath("empty-tfstate.json")
	if err != nil {
		t.Fatalf("Failed to resolve empty-tfstate.json path: %v", err)
	}
	emptyData, _ := json.Marshal(emptyState)
	if err := os.WriteFile(emptyStateFile, emptyData, 0644); err != nil {
		t.Fatalf("Failed to write empty-tfstate.json: %v", err)
	}
	defer os.Remove(emptyStateFile)

	invalidJSONFile, err := getTestDataPath("invalid-tfstate.json")
	if err != nil {
		t.Fatalf("Failed to resolve invalid-tfstate.json path: %v", err)
	}
	if err := os.WriteFile(invalidJSONFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid-tfstate.json: %v", err)
	}
	defer os.Remove(invalidJSONFile)

	nonAWSInstanceState := map[string]interface{}{
		"resources": []interface{}{
			map[string]interface{}{
				"mode":     "managed",
				"type":     "aws_s3_bucket",
				"name":     "example",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"id": "my-bucket",
						},
					},
				},
			},
		},
	}
	nonAWSInstanceFile, err := getTestDataPath("non-aws-instance-tfstate.json")
	if err != nil {
		t.Fatalf("Failed to resolve non-aws-instance-tfstate.json path: %v", err)
	}
	nonAWSData, _ := json.Marshal(nonAWSInstanceState)
	if err := os.WriteFile(nonAWSInstanceFile, nonAWSData, 0644); err != nil {
		t.Fatalf("Failed to write non-aws-instance-tfstate.json: %v", err)
	}
	defer os.Remove(nonAWSInstanceFile)

	tests := []struct {
		name        string
		filePath    string
		expectErr   bool
		expectedLen int
	}{
		{
			name:        "ValidState",
			filePath:    validStateFile,
			expectErr:   false,
			expectedLen: 1,
		},
		{
			name:        "EmptyState",
			filePath:    emptyStateFile,
			expectErr:   false,
			expectedLen: 0,
		},
		{
			name:        "InvalidJSON",
			filePath:    invalidJSONFile,
			expectErr:   true,
			expectedLen: 0,
		},
		{
			name:        "NonAWSInstanceResource",
			filePath:    nonAWSInstanceFile,
			expectErr:   false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger for the test
			logger := logger.NewTestLogger()
			parser := NewTFStateParser(logger)
			configs, err := parser.ParseTFState(tt.filePath)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
			if len(configs) != tt.expectedLen {
				t.Errorf("Expected %d configs, got %d", tt.expectedLen, len(configs))
			}
			if tt.name == "ValidState" && len(configs) > 0 {
				if configs[0].InstanceID != "i-1234567890abcdef0" {
					t.Errorf("Expected instance ID i-1234567890abcdef0, got %s", configs[0].InstanceID)
				}
				if configs[0].InstanceType != "t2.micro" {
					t.Errorf("Expected instance type t2.micro, got %s", configs[0].InstanceType)
				}
				if configs[0].Tags["Name"] != "test-instance" {
					t.Errorf("Expected tag Name=test-instance, got %s", configs[0].Tags["Name"])
				}
			}
		})
	}
}
