package usecases

import (
	"errors"
	"testing"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/interfaces/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for dependencies
type mockAWSClient struct {
	mock.Mock
}

func (m *mockAWSClient) FetchInstanceConfigs() ([]entities.InstanceConfig, error) {
	args := m.Called()
	return args.Get(0).([]entities.InstanceConfig), args.Error(1)
}

type mockTFStateParser struct {
	mock.Mock
}

func (m *mockTFStateParser) ParseTFState(file string) (terraform.InstanceConfigSet, error) {
	args := m.Called(file)
	return args.Get(0).(terraform.InstanceConfigSet), args.Error(1)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *mockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func TestNewDriftDetector(t *testing.T) {
	awsClient := &mockAWSClient{}
	logger := &mockLogger{}
	detector := NewDriftDetector(awsClient, logger)

	assert.NotNil(t, detector)
	assert.Equal(t, awsClient, detector.awsClient)
	assert.Equal(t, logger, detector.logger)
	assert.NotNil(t, detector.tfParser)
}

func TestDetectDrift_Success(t *testing.T) {
	awsClient := &mockAWSClient{}
	tfParser := &mockTFStateParser{}
	logger := &mockLogger{}
	detector := &DriftDetector{
		awsClient: awsClient,
		tfParser:  tfParser,
		logger:    logger,
	}

	// Mock AWS configs
	awsConfigs := []entities.InstanceConfig{
		{
			InstanceID:         "i-123",
			InstanceType:       "t2.micro",
			SubnetID:           "subnet-123",
			SecurityGroupIDs:   []string{"sg-123"},
			IAMInstanceProfile: "profile-123",
			Tags: map[string]string{
				"Name":        "test-instance",
				"Environment": "prod",
			},
			EBSBlockDevices: []entities.EBSBlockDevice{
				{VolumeSize: 8, VolumeType: "gp2"},
			},
		},
	}

	// Mock Terraform configs
	tfConfigs := terraform.InstanceConfigSet{
		InstanceTypes:       []string{"t2.micro"},
		AMIs:                []string{"ami-123"},
		AvailabilityZones:   []string{"us-east-1a"},
		KeyNames:            []string{"key-123"},
		SubnetIDs:           []string{"subnet-123"},
		SecurityGroupIDs:    []string{"sg-123"},
		IAMInstanceProfiles: []string{"profile-123"},
		TagNames:            []string{"test-instance"},
		TagEnvironments:     []string{"prod"},
		EBSVolumeSizes:      []int{8},
		EBSVolumeTypes:      []string{"gp2"},
	}

	// Set up expectations
	awsClient.On("FetchInstanceConfigs").Return(awsConfigs, nil)
	tfParser.On("ParseTFState", "state.tf").Return(tfConfigs, nil)
	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Execute
	err := detector.DetectDrift("state.tf")

	// Assertions
	assert.NoError(t, err)
	awsClient.AssertExpectations(t)
	tfParser.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestDetectDrift_AWSFetchError(t *testing.T) {
	awsClient := &mockAWSClient{}
	tfParser := &mockTFStateParser{}
	logger := &mockLogger{}
	detector := &DriftDetector{
		awsClient: awsClient,
		tfParser:  tfParser,
		logger:    logger,
	}

	// Set up expectations
	awsClient.On("FetchInstanceConfigs").Return([]entities.InstanceConfig{}, errors.New("AWS error"))

	// Execute
	err := detector.DetectDrift("state.tf")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch AWS configs: AWS error")
	awsClient.AssertExpectations(t)
}

func TestDetectDrift_TFParseError(t *testing.T) {
	awsClient := &mockAWSClient{}
	tfParser := &mockTFStateParser{}
	logger := &mockLogger{}
	detector := &DriftDetector{
		awsClient: awsClient,
		tfParser:  tfParser,
		logger:    logger,
	}

	// Mock AWS configs
	awsConfigs := []entities.InstanceConfig{}
	awsClient.On("FetchInstanceConfigs").Return(awsConfigs, nil)
	logger.On("Info", "Fetched AWS configs", mock.Anything).Return()
	tfParser.On("ParseTFState", "state.tf").Return(terraform.InstanceConfigSet{}, errors.New("TF parse error"))

	// Execute
	err := detector.DetectDrift("state.tf")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Terraform state from file state.tf: TF parse error")
	awsClient.AssertExpectations(t)
	tfParser.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestDetectDrift_DriftDetected(t *testing.T) {
	awsClient := &mockAWSClient{}
	tfParser := &mockTFStateParser{}
	logger := &mockLogger{}
	detector := &DriftDetector{
		awsClient: awsClient,
		tfParser:  tfParser,
		logger:    logger,
	}

	// Mock AWS configs with drift
	awsConfigs := []entities.InstanceConfig{
		{
			InstanceID:         "i-123",
			InstanceType:       "t2.micro",
			SubnetID:           "subnet-123",
			SecurityGroupIDs:   []string{"sg-123"},
			IAMInstanceProfile: "profile-123",
			Tags: map[string]string{
				"Name":        "test-instance",
				"Environment": "prod",
			},
			EBSBlockDevices: []entities.EBSBlockDevice{
				{VolumeSize: 8, VolumeType: "gp2"},
			},
		},
	}

	// Mock Terraform configs
	tfConfigs := terraform.InstanceConfigSet{
		InstanceTypes:       []string{"t2.nano"}, // Drift in instance type
		AMIs:                []string{"ami-456"},
		AvailabilityZones:   []string{"us-east-1b"},
		KeyNames:            []string{"key-456"},
		SubnetIDs:           []string{"subnet-456"},     // Drift in subnet
		SecurityGroupIDs:    []string{"sg-456"},         // Drift in security groups
		IAMInstanceProfiles: []string{"profile-456"},    // Drift in IAM profile
		TagNames:            []string{"other-instance"}, // Drift in tags
		TagEnvironments:     []string{"dev"},            // Drift in tags
		EBSVolumeSizes:      []int{16},                  // Drift in EBS size
		EBSVolumeTypes:      []string{"gp3"},            // Drift in EBS type
	}

	// Set up expectations
	awsClient.On("FetchInstanceConfigs").Return(awsConfigs, nil)
	tfParser.On("ParseTFState", "state.tf").Return(tfConfigs, nil)
	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Execute
	err := detector.DetectDrift("state.tf")

	// Assertions
	assert.NoError(t, err)
	awsClient.AssertExpectations(t)
	tfParser.AssertExpectations(t)
	logger.AssertCalled(t, "Info", "Drift detected", mock.Anything)
}

func TestCompareConfigs_NoDrift(t *testing.T) {
	awsConfig := entities.InstanceConfig{
		InstanceID:         "i-123",
		InstanceType:       "t2.micro",
		SubnetID:           "subnet-123",
		SecurityGroupIDs:   []string{"sg-123"},
		IAMInstanceProfile: "profile-123",
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "prod",
		},
		EBSBlockDevices: []entities.EBSBlockDevice{
			{VolumeSize: 8, VolumeType: "gp2"},
		},
	}

	tfConfigs := terraform.InstanceConfigSet{
		InstanceTypes:       []string{"t2.micro"},
		AMIs:                []string{"ami-123"},
		AvailabilityZones:   []string{"us-east-1a"},
		KeyNames:            []string{"key-123"},
		SubnetIDs:           []string{"subnet-123"},
		SecurityGroupIDs:    []string{"sg-123"},
		IAMInstanceProfiles: []string{"profile-123"},
		TagNames:            []string{"test-instance"},
		TagEnvironments:     []string{"prod"},
		EBSVolumeSizes:      []int{8},
		EBSVolumeTypes:      []string{"gp2"},
	}

	diff := compareConfigs(awsConfig, tfConfigs)

	assert.Empty(t, diff)
}

func TestCompareConfigs_WithDrift(t *testing.T) {
	awsConfig := entities.InstanceConfig{
		InstanceID:         "i-123",
		InstanceType:       "t2.micro",
		SubnetID:           "subnet-123",
		SecurityGroupIDs:   []string{"sg-123"},
		IAMInstanceProfile: "profile-123",
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "prod",
		},
		EBSBlockDevices: []entities.EBSBlockDevice{
			{VolumeSize: 8, VolumeType: "gp2"},
		},
	}

	tfConfigs := terraform.InstanceConfigSet{
		InstanceTypes:       []string{"t2.nano"},
		AMIs:                []string{"ami-456"},
		AvailabilityZones:   []string{"us-east-1b"},
		KeyNames:            []string{"key-456"},
		SubnetIDs:           []string{"subnet-456"},
		SecurityGroupIDs:    []string{"sg-456"},
		IAMInstanceProfiles: []string{"profile-456"},
		TagNames:            []string{"other-instance"},
		TagEnvironments:     []string{"dev"},
		EBSVolumeSizes:      []int{16},
		EBSVolumeTypes:      []string{"gp3"},
	}

	diff := compareConfigs(awsConfig, tfConfigs)

	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "instance_type")
	assert.Contains(t, diff, "subnet_id")
	assert.Contains(t, diff, "security_group_ids")
	assert.Contains(t, diff, "iam_instance_profile")
	assert.Contains(t, diff, "tag.Name")
	assert.Contains(t, diff, "tag.Environment")
	assert.Contains(t, diff, "ebs_volume_size")
	assert.Contains(t, diff, "ebs_volume_type")
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, contains(slice, "b"))
	assert.False(t, contains(slice, "d"))
}

func TestContainsInt(t *testing.T) {
	slice := []int{1, 2, 3}
	assert.True(t, containsInt(slice, 2))
	assert.False(t, containsInt(slice, 4))
}

func TestAllIn(t *testing.T) {
	actual := []string{"a", "b"}
	allowed := []string{"a", "b", "c"}
	assert.True(t, allIn(actual, allowed))

	actual = []string{"a", "d"}
	assert.False(t, allIn(actual, allowed))
}

func TestJoinInt(t *testing.T) {
	slice := []int{1, 2, 3}
	result := joinInt(slice)
	assert.Equal(t, "1, 2, 3", result)
}
