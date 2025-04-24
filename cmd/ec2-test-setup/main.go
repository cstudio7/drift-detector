package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file for AWS credentials
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load .env file: %v\n", err)
		os.Exit(1)
	}

	// Create a context
	ctx := context.Background()

	// Create the EC2 client
	ec2Client, err := aws.NewEC2Client(ctx, false) // useMock: false for live requests
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create EC2 client: %v\n", err)
		os.Exit(1)
	}

	// Command-line argument to determine action: "up" or "down"
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [up|down] [instance-id (for down)]")
		fmt.Println("Example: go run main.go up")
		fmt.Println("Example: go run main.go down i-1234567890abcdef0")
		os.Exit(1)
	}

	action := os.Args[1]

	switch action {
	case "up":
		// Parameters for the EC2 instance
		amiID := "ami-0c02fb55956c7d316" // Amazon Linux 2 in us-east-2
		instanceType := "t4g.micro"
		subnetID := "subnet-0fd29464681088be4" // Your subnet ID
		keyName := "test-key"                  // Your key pair name (optional)

		// Create the EC2 instance
		instanceID, err := createEC2Instance(ctx, ec2Client, amiID, instanceType, subnetID, keyName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create EC2 instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("EC2 instance created successfully: %s\n", instanceID)

		// Wait for the instance to be running
		err = waitForInstanceRunning(ctx, ec2Client, instanceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to wait for instance to be running: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("EC2 instance is now running and ready for testing.")
		fmt.Printf("To terminate the instance, run: go run main.go down %s\n", instanceID)

	case "down":
		if len(os.Args) < 3 {
			fmt.Println("Please provide the instance ID to terminate.")
			fmt.Println("Example: go run main.go down i-1234567890abcdef0")
			os.Exit(1)
		}

		instanceID := os.Args[2]

		// Terminate the EC2 instance
		err := ec2Client.TerminateInstance(ctx, instanceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to terminate EC2 instance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("EC2 instance %s terminated successfully.\n", instanceID)

	default:
		fmt.Println("Invalid action. Use 'up' to create an instance or 'down' to terminate one.")
		os.Exit(1)
	}
}

// createEC2Instance creates a new EC2 instance and returns its instance ID.
func createEC2Instance(ctx context.Context, client aws.EC2Client, amiID, instanceType, subnetID, keyName string) (string, error) {
	instanceID, err := client.CreateInstance(ctx, amiID, instanceType, subnetID, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}
	return instanceID, nil
}

// waitForInstanceRunning waits until the EC2 instance is in the "running" state.
func waitForInstanceRunning(ctx context.Context, client aws.EC2Client, instanceID string) error {
	for i := 0; i < 30; i++ { // Retry up to 30 times (5 minutes total)
		instance, err := client.GetInstance(ctx, instanceID)
		if err != nil {
			return fmt.Errorf("failed to describe instance %s: %w", instanceID, err)
		}

		// Check the instance state
		if instance.State != nil && instance.State.Name == "running" {
			return nil
		}

		fmt.Printf("Waiting for instance %s to be running... (attempt %d/30)\n", instanceID, i+1)
		time.Sleep(10 * time.Second) // Wait 10 seconds between checks
	}

	return fmt.Errorf("timeout waiting for instance %s to be running", instanceID)
}
