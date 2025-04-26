package terraform

import (
	"encoding/json"
	"os"
	"testing"

	_ "github.com/cstudio7/drift-detector/internal/interfaces/logger" // Blank identifier to suppress unused import warning
	"github.com/stretchr/testify/assert"
)

// mockLogger is a simple implementation of logger.Logger for testing.
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func TestNewTFStateParser(t *testing.T) {
	mockLog := &mockLogger{}
	parser := NewTFStateParser(mockLog)

	assert.NotNil(t, parser)
	assert.Equal(t, mockLog, parser.logger)
}

func TestParseTFState(t *testing.T) {
	// Create a mock logger
	mockLog := &mockLogger{}

	// Initialize the parser
	parser := NewTFStateParser(mockLog)

	// Test case 1: Successful parsing of a valid Terraform state file
	t.Run("ValidTFStateFile", func(t *testing.T) {
		// Reset mock logger
		mockLog.logs = []string{}

		// Create a valid Terraform state
		validState := TFState{
			Version:          4,
			TerraformVersion: "1.5.0",
			Serial:           1,
			Lineage:          "abc123",
			Resources: struct {
				AWSInstance InstanceConfigSet `json:"aws_instance"`
			}{
				AWSInstance: InstanceConfigSet{
					InstanceTypes:       []string{"t2.micro", "t3.micro"},
					AMIs:                []string{"ami-12345678", "ami-87654321"},
					AvailabilityZones:   []string{"us-east-1a", "us-east-1b"},
					KeyNames:            []string{"key1", "key2"},
					SecurityGroupIDs:    []string{"sg-123", "sg-456"},
					SubnetIDs:           []string{"subnet-123", "subnet-456"},
					IAMInstanceProfiles: []string{"profile1", "profile2"},
					TagNames:            []string{"Name", "Owner"},
					TagEnvironments:     []string{"prod", "dev"},
					EBSVolumeSizes:      []int{10, 20},
					EBSVolumeTypes:      []string{"gp2", "gp3"},
				},
			},
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(validState)
		assert.NoError(t, err)

		// Create a temporary file
		tempFile, err := os.CreateTemp("", "valid_tfstate_*.json")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())

		// Write JSON data to the file
		_, err = tempFile.Write(jsonData)
		assert.NoError(t, err)
		tempFile.Close()

		// Parse the state file
		configSet, err := parser.ParseTFState(tempFile.Name())
		assert.NoError(t, err)

		// Verify the parsed config set
		assert.Equal(t, validState.Resources.AWSInstance, configSet)

		// Verify logging
		expectedLogs := []string{
			"Starting to parse file",
			"Parsed Terraform state",
			"Parsed instance types",
			"Parsed AMIs",
			"Parsed availability zones",
			"Parsed key names",
			"Parsed security groups",
			"Parsed subnet IDs",
			"Parsed IAM instance profiles",
			"Parsed tag names",
			"Parsed tag environments",
			"Parsed EBS volume sizes",
			"Parsed EBS volume types",
			"Returning parsed config set",
		}
		for _, logMsg := range expectedLogs {
			assert.Contains(t, mockLog.logs, logMsg, "Expected log message: %s", logMsg)
		}
	})

	// Test case 2: File does not exist
	t.Run("FileNotFound", func(t *testing.T) {
		// Reset mock logger
		mockLog.logs = []string{}

		// Try to parse a non-existent file
		_, err := parser.ParseTFState("non_existent_file.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")

		// Verify error logging
		assert.Contains(t, mockLog.logs, "Failed to read JSON state file")
	})

	// Test case 3: Invalid JSON content
	t.Run("InvalidJSON", func(t *testing.T) {
		// Reset mock logger
		mockLog.logs = []string{}

		// Create a temporary file with invalid JSON
		tempFile, err := os.CreateTemp("", "invalid_tfstate_*.json")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())

		// Write invalid JSON to the file
		_, err = tempFile.WriteString(`{invalid json}`)
		assert.NoError(t, err)
		tempFile.Close()

		// Parse the state file
		_, err = parser.ParseTFState(tempFile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")

		// Verify logging
		assert.Contains(t, mockLog.logs, "Starting to parse file")
		assert.Contains(t, mockLog.logs, "Failed to parse JSON state file")
	})

	// Test case 4: Empty configurations
	t.Run("EmptyConfig", func(t *testing.T) {
		// Reset mock logger
		mockLog.logs = []string{}

		// Create a Terraform state with empty configurations
		emptyState := TFState{
			Version:          4,
			TerraformVersion: "1.5.0",
			Serial:           1,
			Lineage:          "abc123",
			Resources: struct {
				AWSInstance InstanceConfigSet `json:"aws_instance"`
			}{
				AWSInstance: InstanceConfigSet{
					InstanceTypes:       []string{},
					AMIs:                []string{},
					AvailabilityZones:   []string{},
					KeyNames:            []string{},
					SecurityGroupIDs:    []string{},
					SubnetIDs:           []string{},
					IAMInstanceProfiles: []string{},
					TagNames:            []string{},
					TagEnvironments:     []string{},
					EBSVolumeSizes:      []int{},
					EBSVolumeTypes:      []string{},
				},
			},
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(emptyState)
		assert.NoError(t, err)

		// Create a temporary file
		tempFile, err := os.CreateTemp("", "empty_tfstate_*.json")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())

		// Write JSON data to the file
		_, err = tempFile.Write(jsonData)
		assert.NoError(t, err)
		tempFile.Close()

		// Parse the state file
		configSet, err := parser.ParseTFState(tempFile.Name())
		assert.NoError(t, err)

		// Verify the parsed config set
		assert.Equal(t, emptyState.Resources.AWSInstance, configSet)
		assert.Empty(t, configSet.InstanceTypes)
		assert.Empty(t, configSet.AMIs)
		assert.Empty(t, configSet.AvailabilityZones)
		assert.Empty(t, configSet.KeyNames)
		assert.Empty(t, configSet.SecurityGroupIDs)
		assert.Empty(t, configSet.SubnetIDs)
		assert.Empty(t, configSet.IAMInstanceProfiles)
		assert.Empty(t, configSet.TagNames)
		assert.Empty(t, configSet.TagEnvironments)
		assert.Empty(t, configSet.EBSVolumeSizes)
		assert.Empty(t, configSet.EBSVolumeTypes)

		// Verify logging
		expectedLogs := []string{
			"Starting to parse file",
			"Parsed Terraform state",
			"Parsed instance types",
			"Parsed AMIs",
			"Parsed availability zones",
			"Parsed key names",
			"Parsed security groups",
			"Parsed subnet IDs",
			"Parsed IAM instance profiles",
			"Parsed tag names",
			"Parsed tag environments",
			"Parsed EBS volume sizes",
			"Parsed EBS volume types",
			"Returning parsed config set",
		}
		for _, logMsg := range expectedLogs {
			assert.Contains(t, mockLog.logs, logMsg, "Expected log message: %s", logMsg)
		}
	})
}
