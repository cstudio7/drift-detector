package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

// getTestDataPath resolves the path to the testdata directory dynamically.
func getTestDataPath(filename string) (string, error) {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	// Navigate to the testdata directory at the project root
	// internal/interfaces/terraform/parser_test.go -> testdata/
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(callerFile)))) // Up four levels to project root
	testDataPath := filepath.Join(baseDir, "testdata", filename)
	return testDataPath, nil
}

func TestTFStateParserImpl_ParseTFState(t *testing.T) {
	tests := []struct {
		name      string
		setupFile func(t *testing.T) string
		expectErr bool
		expected  []entities.InstanceConfig
	}{
		{
			name: "ValidState",
			setupFile: func(t *testing.T) string {
				filePath, err := getTestDataPath("sample-tfstate.json")
				if err != nil {
					t.Fatalf("Failed to resolve test data path: %v", err)
				}
				return filePath
			},
			expectErr: false,
			expected: []entities.InstanceConfig{
				{
					InstanceType:       "t2.micro",
					Tags:               map[string]string{"Name": "test-instance"},
					SecurityGroupIDs:   []string{"sg-12345678"},
					SubnetID:           "",
					IAMInstanceProfile: "",
				},
			},
		},
		{
			name: "EmptyState",
			setupFile: func(t *testing.T) string {
				tmpFile, err := os.CreateTemp("", "empty-tfstate.json")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				if err := os.WriteFile(tmpFile.Name(), []byte(`{"resources": []}`), 0644); err != nil {
					t.Fatalf("Failed to write temp file: %v", err)
				}
				return tmpFile.Name()
			},
			expectErr: false,
			expected:  []entities.InstanceConfig{},
		},
		{
			name: "InvalidJSON",
			setupFile: func(t *testing.T) string {
				tmpFile, err := os.CreateTemp("", "invalid-tfstate.json")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				if err := os.WriteFile(tmpFile.Name(), []byte(`{invalid json}`), 0644); err != nil {
					t.Fatalf("Failed to write temp file: %v", err)
				}
				return tmpFile.Name()
			},
			expectErr: true,
		},
		{
			name: "NonAWSInstanceResource",
			setupFile: func(t *testing.T) string {
				tmpFile, err := os.CreateTemp("", "non-aws-tfstate.json")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				content := `{
                    "resources": [
                        {
                            "type": "aws_s3_bucket",
                            "instances": [
                                {
                                    "attributes": {
                                        "bucket": "test-bucket"
                                    }
                                }
                            ]
                        }
                    ]
                }`
				if err := os.WriteFile(tmpFile.Name(), []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write temp file: %v", err)
				}
				return tmpFile.Name()
			},
			expectErr: false,
			expected:  []entities.InstanceConfig{},
		},
	}

	parser := NewTFStateParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			if tt.name != "ValidState" {
				defer os.Remove(filePath)
			}

			configs, err := parser.ParseTFState(filePath)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got error: %v", tt.expectErr, err)
			}
			if err != nil {
				return
			}

			if len(configs) != len(tt.expected) {
				t.Errorf("Expected %d configs, got %d", len(tt.expected), len(configs))
			}
			for i, config := range configs {
				if config.InstanceType != tt.expected[i].InstanceType {
					t.Errorf("Expected InstanceType %q, got %q", tt.expected[i].InstanceType, config.InstanceType)
				}
				if !reflect.DeepEqual(config.Tags, tt.expected[i].Tags) {
					t.Errorf("Expected Tags %v, got %v", tt.expected[i].Tags, config.Tags)
				}
				if !reflect.DeepEqual(config.SecurityGroupIDs, tt.expected[i].SecurityGroupIDs) {
					t.Errorf("Expected SecurityGroupIDs %v, got %v", tt.expected[i].SecurityGroupIDs, config.SecurityGroupIDs)
				}
				if config.SubnetID != tt.expected[i].SubnetID {
					t.Errorf("Expected SubnetID %q, got %q", tt.expected[i].SubnetID, config.SubnetID)
				}
				if config.IAMInstanceProfile != tt.expected[i].IAMInstanceProfile {
					t.Errorf("Expected IAMInstanceProfile %q, got %q", tt.expected[i].IAMInstanceProfile, config.IAMInstanceProfile)
				}
			}
		})
	}
}
