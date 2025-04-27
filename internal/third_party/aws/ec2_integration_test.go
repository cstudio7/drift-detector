package aws

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// TestEC2ClientImpl_CreateAndGetInstance_Integration is an integration test that creates a real EC2 instance.
func TestEC2ClientImpl_CreateAndGetInstance_Integration(t *testing.T) {
	// Skip this test unless explicitly run (e.g., with a flag or environment variable)
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test; set RUN_INTEGRATION_TESTS=true to run")
	}

	// Load the .env file
	if err := godotenv.Load(); err != nil {
		t.Logf("Failed to load .env file (optional for tests): %v", err)
	}

	ctx := context.Background()
	client, err := NewEC2Client(ctx, false) // useMock: false for live AWS requests
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	// Replace these with valid values from your AWS account
	amiID := "ami-0430580de6244e02e" // Amazon Linux 2 in us-east-2
	instanceType := "t2.micro"
	subnetID := "subnet-0fd29464681088be4" // Replace with a valid subnet ID in your VPC
	keyName := "test-key"                  // Replace with your key pair name (optional)

	// Create the instance
	instanceID, err := client.CreateInstance(ctx, amiID, instanceType, subnetID, keyName)
	if err != nil {
		t.Fatalf("Failed to create EC2 instance: %v", err)
	}
	t.Logf("Created EC2 instance: %s", instanceID)

	// Cleanup: Terminate the instance after the test
	defer func() {
		err := client.TerminateInstance(ctx, instanceID)
		if err != nil {
			t.Logf("Failed to terminate EC2 instance %s: %v", instanceID, err)
		} else {
			t.Logf("Terminated EC2 instance: %s", instanceID)
		}
	}()

	// Wait for the instance to be running (optional, depending on your needs)
	time.Sleep(30 * time.Second) // Wait for the instance to start (adjust as needed)

	// Fetch the instance
	instance, err := client.GetInstance(ctx, instanceID)
	if err != nil {
		t.Fatalf("Failed to get EC2 instance: %v", err)
	}

	// Verify the instance details
	if *instance.InstanceId != instanceID {
		t.Errorf("Expected instance ID %s, got %s", instanceID, *instance.InstanceId)
	}
	if string(instance.InstanceType) != instanceType {
		t.Errorf("Expected instance type %s, got %s", instanceType, instance.InstanceType)
	}
}
