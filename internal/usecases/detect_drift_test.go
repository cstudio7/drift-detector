package usecases

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/domain/services"
)

// getTestDataPath resolves the path to the testdata directory dynamically.
func getTestDataPath(filename string) (string, error) {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	// Navigate to the testdata directory at the project root
	// internal/usecases/detect_drift_test.go -> testdata/
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(callerFile))) // Up three levels to project root
	testDataPath := filepath.Join(baseDir, "testdata", filename)
	return testDataPath, nil
}

// MockEC2Client mocks the EC2Client interface.
type MockEC2Client struct {
	GetInstanceFunc      func(ctx context.Context, instanceID string) (*awstypes.Instance, error)
	ToInstanceConfigFunc func(instance *awstypes.Instance) entities.InstanceConfig
}

func (m *MockEC2Client) GetInstance(ctx context.Context, instanceID string) (*awstypes.Instance, error) {
	return m.GetInstanceFunc(ctx, instanceID)
}

func (m *MockEC2Client) ToInstanceConfig(instance *awstypes.Instance) entities.InstanceConfig {
	return m.ToInstanceConfigFunc(instance)
}

// MockTFStateParser mocks the TFStateParser interface.
type MockTFStateParser struct {
	ParseTFStateFunc func(filePath string) ([]entities.InstanceConfig, error)
}

func (m *MockTFStateParser) ParseTFState(filePath string) ([]entities.InstanceConfig, error) {
	return m.ParseTFStateFunc(filePath)
}

// MockLogger mocks the Logger interface.
type MockLogger struct {
	InfoFunc   func(msg string, keysAndValues ...interface{})
	WarnFunc   func(msg string, keysAndValues ...interface{})
	ErrorFunc  func(msg string, keysAndValues ...interface{})
	driftCount int
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	if msg == "Drift detected" {
		m.driftCount++
	}
	m.InfoFunc(msg, keysAndValues...)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.WarnFunc(msg, keysAndValues...)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.ErrorFunc(msg, keysAndValues...)
}

func (m *MockLogger) Cleanup(t *testing.T, expectedDriftCount int) {
	if m.driftCount != expectedDriftCount {
		t.Errorf("Expected %d drift detections, got %d", expectedDriftCount, m.driftCount)
	}
}

func TestDetectDriftUseCase_Execute(t *testing.T) {
	driftService := services.NewDriftService()

	// Mock AWS instance
	awsInstance := &awstypes.Instance{
		InstanceId:   stringPtr("i-1234567890abcdef0"),
		InstanceType: "t2.small",
		Tags: []awstypes.Tag{
			{Key: stringPtr("Name"), Value: stringPtr("prod-instance")},
		},
		SecurityGroups: []awstypes.GroupIdentifier{
			{GroupId: stringPtr("sg-12345678")},
		},
	}
	awsConfig := entities.InstanceConfig{
		InstanceID:       "i-1234567890abcdef0",
		InstanceType:     "t2.small",
		Tags:             map[string]string{"Name": "prod-instance"},
		SecurityGroupIDs: []string{"sg-12345678"},
	}

	// Second mock AWS instance
	awsInstance2 := &awstypes.Instance{
		InstanceId:   stringPtr("i-0987654321fedcba0"),
		InstanceType: "t2.medium",
		Tags: []awstypes.Tag{
			{Key: stringPtr("Name"), Value: stringPtr("prod-instance-2")},
		},
		SecurityGroups: []awstypes.GroupIdentifier{
			{GroupId: stringPtr("sg-87654321")},
		},
	}
	awsConfig2 := entities.InstanceConfig{
		InstanceID:       "i-0987654321fedcba0",
		InstanceType:     "t2.medium",
		Tags:             map[string]string{"Name": "prod-instance-2"},
		SecurityGroupIDs: []string{"sg-87654321"},
	}

	// Terraform config
	tfConfigs := []entities.InstanceConfig{
		{
			InstanceID:       "i-1234567890abcdef0",
			InstanceType:     "t2.micro",
			Tags:             map[string]string{"Name": "test-instance"},
			SecurityGroupIDs: []string{"sg-12345678"},
		},
		{
			InstanceID:       "i-0987654321fedcba0",
			InstanceType:     "t2.micro",
			Tags:             map[string]string{"Name": "test-instance-2"},
			SecurityGroupIDs: []string{"sg-87654321"},
		},
	}

	// Resolve the path to sample-tfstate.json dynamically
	tfStatePath, err := getTestDataPath("sample-tfstate.json")
	if err != nil {
		t.Fatalf("Failed to resolve test data path: %v", err)
	}

	tests := []struct {
		name               string
		instanceIDs        []string
		tfStatePath        string
		attributes         []string
		mockAWS            func() *MockEC2Client
		mockTFParser       func() *MockTFStateParser
		mockLogger         func(t *testing.T) *MockLogger
		expectErr          bool
		expectedLogs       []string
		expectedDriftCount int
	}{
		{
			name:        "SuccessWithDrift",
			instanceIDs: []string{"i-1234567890abcdef0"},
			tfStatePath: tfStatePath,
			attributes:  []string{"instance_type", "tags", "security_groups"},
			mockAWS: func() *MockEC2Client {
				return &MockEC2Client{
					GetInstanceFunc: func(ctx context.Context, instanceID string) (*awstypes.Instance, error) {
						return awsInstance, nil
					},
					ToInstanceConfigFunc: func(instance *awstypes.Instance) entities.InstanceConfig {
						return awsConfig
					},
				}
			},
			mockTFParser: func() *MockTFStateParser {
				return &MockTFStateParser{
					ParseTFStateFunc: func(filePath string) ([]entities.InstanceConfig, error) {
						return tfConfigs, nil
					},
				}
			},
			mockLogger: func(t *testing.T) *MockLogger {
				return &MockLogger{
					InfoFunc: func(msg string, keysAndValues ...interface{}) {
						if msg == "Drift detected" {
							changes := keysAndValues[3].(map[string]entities.Change)
							if changes["instance_type"].Expected != "t2.micro" || changes["instance_type"].Actual != "t2.small" {
								t.Errorf("Expected instance_type drift, got %v", changes["instance_type"])
							}
							if changes["tags"].Expected.(map[string]string)["Name"] != "test-instance" || changes["tags"].Actual.(map[string]string)["Name"] != "prod-instance" {
								t.Errorf("Expected tags drift, got %v", changes["tags"])
							}
						}
					},
					WarnFunc:  func(msg string, keysAndValues ...interface{}) {},
					ErrorFunc: func(msg string, keysAndValues ...interface{}) {},
				}
			},
			expectErr:          false,
			expectedLogs:       []string{"Drift detected"},
			expectedDriftCount: 1,
		},
		{
			name:        "InstanceNotFound",
			instanceIDs: []string{"i-nonexistent"},
			tfStatePath: tfStatePath,
			attributes:  []string{"instance_type"},
			mockAWS: func() *MockEC2Client {
				return &MockEC2Client{
					GetInstanceFunc: func(ctx context.Context, instanceID string) (*awstypes.Instance, error) {
						return nil, entities.ErrInstanceNotFound
					},
					ToInstanceConfigFunc: func(instance *awstypes.Instance) entities.InstanceConfig {
						return entities.InstanceConfig{}
					},
				}
			},
			mockTFParser: func() *MockTFStateParser {
				return &MockTFStateParser{
					ParseTFStateFunc: func(filePath string) ([]entities.InstanceConfig, error) {
						return tfConfigs, nil
					},
				}
			},
			mockLogger: func(t *testing.T) *MockLogger {
				return &MockLogger{
					InfoFunc: func(msg string, keysAndValues ...interface{}) {},
					WarnFunc: func(msg string, keysAndValues ...interface{}) {
						if msg != "Instance not found in AWS" {
							t.Errorf("Expected warn log 'Instance not found in AWS', got %s", msg)
						}
					},
					ErrorFunc: func(msg string, keysAndValues ...interface{}) {},
				}
			},
			expectErr:          false,
			expectedLogs:       []string{"Instance not found in AWS"},
			expectedDriftCount: 0,
		},
		{
			name:        "TFStateParseError",
			instanceIDs: []string{"i-1234567890abcdef0"},
			tfStatePath: "nonexistent-tfstate.json",
			attributes:  []string{"instance_type"},
			mockAWS: func() *MockEC2Client {
				return &MockEC2Client{
					GetInstanceFunc: func(ctx context.Context, instanceID string) (*awstypes.Instance, error) {
						return awsInstance, nil
					},
					ToInstanceConfigFunc: func(instance *awstypes.Instance) entities.InstanceConfig {
						return awsConfig
					},
				}
			},
			mockTFParser: func() *MockTFStateParser {
				return &MockTFStateParser{
					ParseTFStateFunc: func(filePath string) ([]entities.InstanceConfig, error) {
						return nil, errors.New("parse error")
					},
				}
			},
			mockLogger: func(t *testing.T) *MockLogger {
				return &MockLogger{
					InfoFunc: func(msg string, keysAndValues ...interface{}) {},
					WarnFunc: func(msg string, keysAndValues ...interface{}) {},
					ErrorFunc: func(msg string, keysAndValues ...interface{}) {
						if msg != "Failed to parse Terraform state" {
							t.Errorf("Expected error log 'Failed to parse Terraform state', got %s", msg)
						}
					},
				}
			},
			expectErr:          true,
			expectedLogs:       []string{"Failed to parse Terraform state"},
			expectedDriftCount: 0,
		},
		{
			name:        "MultipleInstances",
			instanceIDs: []string{"i-1234567890abcdef0", "i-0987654321fedcba0"},
			tfStatePath: tfStatePath,
			attributes:  []string{"instance_type", "tags", "security_groups"},
			mockAWS: func() *MockEC2Client {
				return &MockEC2Client{
					GetInstanceFunc: func(ctx context.Context, instanceID string) (*awstypes.Instance, error) {
						if instanceID == "i-1234567890abcdef0" {
							return awsInstance, nil
						}
						if instanceID == "i-0987654321fedcba0" {
							return awsInstance2, nil
						}
						return nil, entities.ErrInstanceNotFound
					},
					ToInstanceConfigFunc: func(instance *awstypes.Instance) entities.InstanceConfig {
						if *instance.InstanceId == "i-1234567890abcdef0" {
							return awsConfig
						}
						return awsConfig2
					},
				}
			},
			mockTFParser: func() *MockTFStateParser {
				return &MockTFStateParser{
					ParseTFStateFunc: func(filePath string) ([]entities.InstanceConfig, error) {
						return tfConfigs, nil
					},
				}
			},
			mockLogger: func(t *testing.T) *MockLogger {
				return &MockLogger{
					InfoFunc: func(msg string, keysAndValues ...interface{}) {
						if msg == "Drift detected" {
							instanceID := keysAndValues[1].(string)
							changes := keysAndValues[3].(map[string]entities.Change)
							if instanceID == "i-1234567890abcdef0" {
								if changes["instance_type"].Expected != "t2.micro" || changes["instance_type"].Actual != "t2.small" {
									t.Errorf("Expected instance_type drift for i-1234567890abcdef0, got %v", changes["instance_type"])
								}
								if changes["tags"].Expected.(map[string]string)["Name"] != "test-instance" || changes["tags"].Actual.(map[string]string)["Name"] != "prod-instance" {
									t.Errorf("Expected tags drift for i-1234567890abcdef0, got %v", changes["tags"])
								}
							} else if instanceID == "i-0987654321fedcba0" {
								if changes["instance_type"].Expected != "t2.micro" || changes["instance_type"].Actual != "t2.medium" {
									t.Errorf("Expected instance_type drift for i-0987654321fedcba0, got %v", changes["instance_type"])
								}
								if changes["tags"].Expected.(map[string]string)["Name"] != "test-instance-2" || changes["tags"].Actual.(map[string]string)["Name"] != "prod-instance-2" {
									t.Errorf("Expected tags drift for i-0987654321fedcba0, got %v", changes["tags"])
								}
							}
						}
					},
					WarnFunc:  func(msg string, keysAndValues ...interface{}) {},
					ErrorFunc: func(msg string, keysAndValues ...interface{}) {},
				}
			},
			expectErr:          false,
			expectedLogs:       []string{"Drift detected", "Drift detected"},
			expectedDriftCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := tt.mockLogger(t)
			useCase := NewDetectDriftUseCase(
				driftService,
				tt.mockAWS(),
				tt.mockTFParser(),
				mockLogger,
				tt.instanceIDs,
				tt.tfStatePath,
				tt.attributes,
			)

			err := useCase.Execute(context.Background())
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got error: %v", tt.expectErr, err)
			}
			mockLogger.Cleanup(t, tt.expectedDriftCount)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
