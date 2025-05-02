package usecases

import (
	"errors"
	"testing"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
	"github.com/stretchr/testify/assert"
)

// Mocks

type mockAWSClient struct {
	fetchConfigs func() ([]entities.InstanceConfig, error)
}

func (m *mockAWSClient) FetchInstanceConfigs() ([]entities.InstanceConfig, error) {
	return m.fetchConfigs()
}

type mockTFStateParser struct {
	parseFunc func(string) (terraform.InstanceConfigSet, error)
}

func (m *mockTFStateParser) ParseTFState(tfStateFile string) (terraform.InstanceConfigSet, error) {
	return m.parseFunc(tfStateFile)
}

type mockLogger struct{}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {}

// Helpers

func newDriftDetector(awsClient aws.AWSClient, tfParser terraform.TFStateParser, logger logger.Logger) *DriftDetector {
	detector := NewDriftDetector(awsClient, logger)
	detector.tfParser = tfParser
	return detector
}

// Tests

func TestNewDriftDetector(t *testing.T) {
	mockAWS := &mockAWSClient{}
	mockLogger := &mockLogger{}
	detector := NewDriftDetector(mockAWS, mockLogger)

	assert.Equal(t, mockAWS, detector.AWSClient())
	assert.NotNil(t, detector.TFParser())
	assert.Equal(t, mockLogger, detector.Logger())
}

func TestDetectDrift_Success_NoDrift(t *testing.T) {
	mockAWS := &mockAWSClient{
		fetchConfigs: func() ([]entities.InstanceConfig, error) {
			return []entities.InstanceConfig{
				{
					InstanceID:         "i-12345",
					InstanceType:       "t2.micro",
					SecurityGroupIDs:   []string{"sg-123"},
					SubnetID:           "subnet-123",
					IAMInstanceProfile: "profile-123",
					Tags: map[string]string{
						"Name":        "test-instance",
						"Environment": "dev",
					},
					EBSBlockDevices: []entities.EBSBlockDevice{
						{VolumeSize: 8, VolumeType: "gp2"},
					},
				},
			}, nil
		},
	}

	mockTF := &mockTFStateParser{
		parseFunc: func(tfStateFile string) (terraform.InstanceConfigSet, error) {
			return terraform.InstanceConfigSet{
				InstanceTypes:       []string{"t2.micro"},
				SecurityGroupIDs:    []string{"sg-123"},
				SubnetIDs:           []string{"subnet-123"},
				IAMInstanceProfiles: []string{"profile-123"},
				TagNames:            []string{"test-instance"},
				TagEnvironments:     []string{"dev"},
				EBSVolumeSizes:      []int{8},
				EBSVolumeTypes:      []string{"gp2"},
			}, nil
		},
	}

	logger := &mockLogger{}
	detector := newDriftDetector(mockAWS, mockTF, logger)

	err := detector.DetectDrift("mock.tfstate")
	assert.NoError(t, err)
}

func TestDetectDrift_FetchAWSConfigsError(t *testing.T) {
	mockAWS := &mockAWSClient{
		fetchConfigs: func() ([]entities.InstanceConfig, error) {
			return nil, errors.New("AWS fetch error")
		},
	}

	mockTF := &mockTFStateParser{}
	logger := &mockLogger{}
	detector := newDriftDetector(mockAWS, mockTF, logger)

	err := detector.DetectDrift("mock.tfstate")
	assert.ErrorIs(t, err, entities.ErrFetchAWSConfigs)
}

func TestDetectDrift_EmptyAWSConfigs(t *testing.T) {
	mockAWS := &mockAWSClient{
		fetchConfigs: func() ([]entities.InstanceConfig, error) {
			return []entities.InstanceConfig{}, nil
		},
	}

	mockTF := &mockTFStateParser{}
	logger := &mockLogger{}
	detector := newDriftDetector(mockAWS, mockTF, logger)

	err := detector.DetectDrift("mock.tfstate")
	assert.ErrorIs(t, err, entities.ErrEmptyConfigs)
}

func TestDetectDrift_ParseTFStateError(t *testing.T) {
	mockAWS := &mockAWSClient{
		fetchConfigs: func() ([]entities.InstanceConfig, error) {
			return []entities.InstanceConfig{
				{InstanceID: "i-12345"},
			}, nil
		},
	}

	mockTF := &mockTFStateParser{
		parseFunc: func(tfStateFile string) (terraform.InstanceConfigSet, error) {
			return terraform.InstanceConfigSet{}, errors.New("parse error")
		},
	}

	logger := &mockLogger{}
	detector := newDriftDetector(mockAWS, mockTF, logger)

	err := detector.DetectDrift("mock.tfstate")
	assert.ErrorIs(t, err, entities.ErrInvalidTerraformState)
}

func TestDetectDrift_EmptyTFConfigs(t *testing.T) {
	mockAWS := &mockAWSClient{
		fetchConfigs: func() ([]entities.InstanceConfig, error) {
			return []entities.InstanceConfig{
				{InstanceID: "i-12345"},
			}, nil
		},
	}

	mockTF := &mockTFStateParser{
		parseFunc: func(tfStateFile string) (terraform.InstanceConfigSet, error) {
			return terraform.InstanceConfigSet{}, nil
		},
	}

	logger := &mockLogger{}
	detector := newDriftDetector(mockAWS, mockTF, logger)

	err := detector.DetectDrift("mock.tfstate")
	assert.ErrorIs(t, err, entities.ErrEmptyConfigs)
}

func TestDetectDrift_WithDriftDetected(t *testing.T) {
	mockAWS := &mockAWSClient{
		fetchConfigs: func() ([]entities.InstanceConfig, error) {
			return []entities.InstanceConfig{
				{
					InstanceID:         "i-67890",
					InstanceType:       "t3.medium",
					SecurityGroupIDs:   []string{"sg-456"},
					SubnetID:           "subnet-456",
					IAMInstanceProfile: "profile-456",
					Tags: map[string]string{
						"Name":        "drift-instance",
						"Environment": "prod",
					},
					EBSBlockDevices: []entities.EBSBlockDevice{
						{VolumeSize: 20, VolumeType: "gp3"},
					},
				},
			}, nil
		},
	}

	mockTF := &mockTFStateParser{
		parseFunc: func(tfStateFile string) (terraform.InstanceConfigSet, error) {
			return terraform.InstanceConfigSet{
				InstanceTypes:       []string{"t2.micro"},
				SecurityGroupIDs:    []string{"sg-123"},
				SubnetIDs:           []string{"subnet-123"},
				IAMInstanceProfiles: []string{"profile-123"},
				TagNames:            []string{"test-instance"},
				TagEnvironments:     []string{"dev"},
				EBSVolumeSizes:      []int{8},
				EBSVolumeTypes:      []string{"gp2"},
			}, nil
		},
	}

	logger := &mockLogger{}
	detector := newDriftDetector(mockAWS, mockTF, logger)

	err := detector.DetectDrift("mock.tfstate")
	assert.NoError(t, err)
}
