package aws

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cstudio7/drift-detector/internal/domain/entities"
)

func TestEC2ClientImpl_GetInstance(t *testing.T) {
	// Create a mock EC2 instance response
	mockInstance := types.Instance{
		InstanceId:   aws.String("i-1234567890abcdef0"),
		InstanceType: "t2.micro",
		Tags: []types.Tag{
			{Key: aws.String("Name"), Value: aws.String("test-instance")},
		},
		SecurityGroups: []types.GroupIdentifier{
			{GroupId: aws.String("sg-12345678")},
		},
		SubnetId: aws.String("subnet-12345678"),
	}

	// Write mock data to a file
	mockResp := struct {
		Reservations []struct {
			Instances []types.Instance
		}
	}{
		Reservations: []struct {
			Instances []types.Instance
		}{
			{Instances: []types.Instance{mockInstance}},
		},
	}
	mockData, _ := json.Marshal(mockResp)
	filePath, _ := getTestDataPath("sample-ec2.json")
	os.WriteFile(filePath, mockData, 0644)
	defer os.Remove(filePath)

	// Create the EC2 client with mock enabled
	ctx := context.Background()
	client, err := NewEC2Client(ctx, true) // useMock: true
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	// Test successful retrieval
	instance, err := client.GetInstance(ctx, "i-1234567890abcdef0")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if *instance.InstanceId != "i-1234567890abcdef0" {
		t.Errorf("Expected instance ID i-1234567890abcdef0, got %s", *instance.InstanceId)
	}

	// Test instance not found
	_, err = client.GetInstance(ctx, "i-nonexistent")
	if err != entities.ErrInstanceNotFound {
		t.Errorf("Expected ErrInstanceNotFound, got: %v", err)
	}
}

func TestEC2ClientImpl_ToInstanceConfig(t *testing.T) {
	// Create a mock EC2 instance
	instance := &types.Instance{
		InstanceType: "t2.micro",
		Tags: []types.Tag{
			{Key: aws.String("Name"), Value: aws.String("test-instance")},
		},
		SecurityGroups: []types.GroupIdentifier{
			{GroupId: aws.String("sg-12345678")},
		},
		SubnetId: aws.String("subnet-12345678"),
		IamInstanceProfile: &types.IamInstanceProfile{
			Arn: aws.String("arn:aws:iam::123456789012:instance-profile/test-profile"),
		},
	}

	// Create the EC2 client
	ctx := context.Background()
	client, err := NewEC2Client(ctx, true) // useMock: true
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	// Convert to InstanceConfig
	config := client.ToInstanceConfig(instance)
	if config.InstanceType != "t2.micro" {
		t.Errorf("Expected InstanceType t2.micro, got %s", config.InstanceType)
	}
	if config.Tags["Name"] != "test-instance" {
		t.Errorf("Expected Tags[Name] test-instance, got %s", config.Tags["Name"])
	}
	if len(config.SecurityGroupIDs) != 1 || config.SecurityGroupIDs[0] != "sg-12345678" {
		t.Errorf("Expected SecurityGroupIDs [sg-12345678], got %v", config.SecurityGroupIDs)
	}
	if config.SubnetID != "subnet-12345678" {
		t.Errorf("Expected SubnetID subnet-12345678, got %s", config.SubnetID)
	}
	if config.IAMInstanceProfile != "arn:aws:iam::123456789012:instance-profile/test-profile" {
		t.Errorf("Expected IAMInstanceProfile arn:aws:iam::123456789012:instance-profile/test-profile, got %s", config.IAMInstanceProfile)
	}
}
