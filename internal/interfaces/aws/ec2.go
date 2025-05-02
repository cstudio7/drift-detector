package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"encoding/json"
	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/pkg/aws"
)

// getTestDataPath resolves the path to the testdata directory dynamically.
func getTestDataPath(filename string) (string, error) {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	// Navigate to the testdata directory at the project root
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(callerFile)))) // Up four levels to project root
	testDataPath := filepath.Join(baseDir, "testdata", filename)
	return testDataPath, nil
}

// EC2Client defines the interface for interacting with AWS EC2.
type EC2Client interface {
	GetInstance(ctx context.Context, instanceID string) (*aws.Instance, error)
	CreateInstance(ctx context.Context, amiID, instanceType, subnetID, keyName string) (string, error)
	TerminateInstance(ctx context.Context, instanceID string) error
	ToInstanceConfig(instance *aws.Instance) entities.InstanceConfig
	Client() *aws.EC2Client // Added Client method to the interface
}

// EC2ClientImpl is the implementation of EC2Client.
type EC2ClientImpl struct {
	client  *aws.EC2Client
	useMock bool
}

// NewEC2Client creates a new EC2ClientImpl.
func NewEC2Client(ctx context.Context, useMock bool) (*EC2ClientImpl, error) {
	client, err := aws.Initialize(ctx) // Ensure AWS initialize method is defined in your code
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS client: %w", err)
	}

	return &EC2ClientImpl{
		client:  client,
		useMock: useMock,
	}, nil
}

// CreateInstance creates a new EC2 instance and returns its instance ID.
func (c *EC2ClientImpl) CreateInstance(ctx context.Context, amiID, instanceType, subnetID, keyName string) (string, error) {
	if c.useMock {
		// Mock implementation: return a fake instance ID
		return "i-mock1234567890abcdef0", nil
	}

	// Live AWS request to create an instance
	input := &aws.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: aws.InstanceType(instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		SubnetId:     aws.String(subnetID),
		TagSpecifications: []aws.TagSpecification{
			{
				ResourceType: aws.ResourceTypeInstance,
				Tags: []aws.Tag{
					{Key: aws.String("Name"), Value: aws.String("test-instance")},
				},
			},
		},
	}

	// Set the key pair if provided
	if keyName != "" {
		input.KeyName = aws.String(keyName)
	}

	result, err := c.client.RunInstances(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create EC2 instance: %w", err)
	}

	if len(result.Instances) == 0 {
		return "", fmt.Errorf("no instances created")
	}

	instanceID := *result.Instances[0].InstanceId
	return instanceID, nil
}

// TerminateInstance terminates an EC2 instance.
func (c *EC2ClientImpl) TerminateInstance(ctx context.Context, instanceID string) error {
	if c.useMock {
		// Mock implementation: do nothing
		return nil
	}

	// Live AWS request to terminate the instance
	input := &aws.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := c.client.TerminateInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to terminate EC2 instance %s: %w", instanceID, err)
	}

	return nil
}

// GetInstance retrieves an EC2 instance by ID.
func (c *EC2ClientImpl) GetInstance(ctx context.Context, instanceID string) (*aws.Instance, error) {
	if c.useMock {
		// Mock implementation for testing
		filePath, err := getTestDataPath("sample-ec2.json")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve test data path: %w", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read mock data: %w", err)
		}

		var resp struct {
			Reservations []struct {
				Instances []aws.Instance
			}
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse mock data: %w", err)
		}

		if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			return nil, entities.ErrInstanceNotFound
		}

		if resp.Reservations[0].Instances[0].InstanceId == nil || *resp.Reservations[0].Instances[0].InstanceId != instanceID {
			return nil, entities.ErrInstanceNotFound
		}

		return &resp.Reservations[0].Instances[0], nil
	}

	// Live AWS request
	input := &aws.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	result, err := c.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance %s: %w", instanceID, err)
	}

	// Check if the instance was found
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId != nil && *instance.InstanceId == instanceID {
				return &instance, nil
			}
		}
	}

	return nil, entities.ErrInstanceNotFound
}

// ToInstanceConfig converts an EC2 instance to an InstanceConfig struct.
func (c *EC2ClientImpl) ToInstanceConfig(instance *aws.Instance) entities.InstanceConfig {
	config := entities.InstanceConfig{
		InstanceID:         aws.ToString(instance.InstanceId), // Correct usage of aws.ToString
		InstanceType:       string(instance.InstanceType),     // Correct reference to instance type
		Tags:               make(map[string]string),
		SecurityGroupIDs:   make([]string, 0),
		SubnetID:           "",
		IAMInstanceProfile: "",
	}
	for _, tag := range instance.Tags {
		if tag.Key != nil && tag.Value != nil {
			config.Tags[*tag.Key] = *tag.Value
		}
	}
	for _, sg := range instance.SecurityGroups {
		if sg.GroupId != nil {
			config.SecurityGroupIDs = append(config.SecurityGroupIDs, *sg.GroupId)
		}
	}
	if instance.SubnetId != nil {
		config.SubnetID = *instance.SubnetId
	}
	if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
		config.IAMInstanceProfile = *instance.IamInstanceProfile.Arn
	}
	return config
}

// Client returns the underlying aws.EC2Client for direct access.
func (c *EC2ClientImpl) Client() *aws.EC2Client {
	return c.client
}
