package usecases

import (
	"testing"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

// MockAWSClient is a mock implementation of aws.AWSClient.
type MockAWSClient struct {
	configs []entities.InstanceConfig
	err     error
}

func (m *MockAWSClient) FetchInstanceConfigs() ([]entities.InstanceConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configs, nil
}

// MockTFStateParser is a mock implementation of terraform.TFStateParser.
type MockTFStateParser struct {
	configs []entities.InstanceConfig
	err     error
}

func (m *MockTFStateParser) ParseTFState(_ string) ([]entities.InstanceConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configs, nil
}

// MockLogger is a mock implementation of logger.Logger.
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func TestDriftDetector_DetectDrift(t *testing.T) {
	tests := []struct {
		name           string
		awsConfigs     []entities.InstanceConfig
		tfConfigs      []entities.InstanceConfig
		expectedDrifts int
	}{
		{
			name: "NoDrift",
			awsConfigs: []entities.InstanceConfig{
				{
					InstanceID:   "i-123",
					InstanceType: "t2.micro",
					Tags:         map[string]string{"Name": "test"},
				},
			},
			tfConfigs: []entities.InstanceConfig{
				{
					InstanceID:   "i-123",
					InstanceType: "t2.micro",
					Tags:         map[string]string{"Name": "test"},
				},
			},
			expectedDrifts: 0,
		},
		{
			name: "DriftDetected",
			awsConfigs: []entities.InstanceConfig{
				{
					InstanceID:   "i-123",
					InstanceType: "t2.micro",
					Tags:         map[string]string{"Name": "test"},
				},
			},
			tfConfigs: []entities.InstanceConfig{
				{
					InstanceID:   "i-123",
					InstanceType: "t2.small", // Different instance type
					Tags:         map[string]string{"Name": "test"},
				},
			},
			expectedDrifts: 1,
		},
		{
			name: "InstanceNotInTerraform",
			awsConfigs: []entities.InstanceConfig{
				{
					InstanceID:   "i-123",
					InstanceType: "t2.micro",
				},
			},
			tfConfigs:      []entities.InstanceConfig{},
			expectedDrifts: 1,
		},
		{
			name:       "InstanceNotInAWS",
			awsConfigs: []entities.InstanceConfig{},
			tfConfigs: []entities.InstanceConfig{
				{
					InstanceID:   "i-123",
					InstanceType: "t2.micro",
				},
			},
			expectedDrifts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			awsClient := &MockAWSClient{configs: tt.awsConfigs}
			tfParser := &MockTFStateParser{configs: tt.tfConfigs}
			logger := &MockLogger{}

			// Create DriftDetector
			detector := &DriftDetector{
				awsClient: awsClient,
				tfParser:  tfParser,
				logger:    logger,
			}

			// Run DetectDrift
			err := detector.DetectDrift("dummy.tfstate")
			if err != nil {
				t.Fatalf("DetectDrift failed: %v", err)
			}

			// Count the number of drift logs
			driftCount := 0
			for _, log := range logger.logs {
				if log == "Drift detected" || log == "Drift detected: Instance not found in Terraform state" || log == "Drift detected: Instance not found in AWS" {
					driftCount++
				}
			}

			if driftCount != tt.expectedDrifts {
				t.Errorf("Expected %d drifts, got %d", tt.expectedDrifts, driftCount)
			}
		})
	}
}

func TestNewDriftDetector(t *testing.T) {
	awsClient := &MockAWSClient{}
	logger := &MockLogger{}

	detector := NewDriftDetector(awsClient, logger)

	if detector.awsClient != awsClient {
		t.Error("Expected awsClient to be set correctly")
	}
	// Compare logger by checking logs instead of direct comparison
	detector.logger.Info("test")
	if len(logger.logs) != 1 || logger.logs[0] != "test" {
		t.Error("Expected logger to be set correctly")
	}
	if detector.tfParser == nil {
		t.Error("Expected tfParser to be initialized")
	}
}
