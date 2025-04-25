package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	customaws "github.com/cstudio7/drift-detector/internal/interfaces/aws"
	"github.com/cstudio7/drift-detector/internal/interfaces/logger"
	"github.com/cstudio7/drift-detector/internal/usecases"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load .env file: %v\n", err)
		os.Exit(1)
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Command-line argument to determine action: "up", "down", or "detect"
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [up|down|detect] [instance-id (for down)] [tfstate-file (for detect)]")
		fmt.Println("Example: go run main.go up")
		fmt.Println("Example: go run main.go down i-1234567890abcdef0")
		fmt.Println("Example: go run main.go detect terraform.tfstate")
		os.Exit(1)
	}

	action := os.Args[1]

	// Create a context
	ctx := context.Background()

	// Create a logger
	logger := logger.NewStdLogger()

	switch action {
	case "up":
		// Create the EC2 client for setup/teardown
		ec2Client, err := customaws.NewEC2Client(ctx, false)
		if err != nil {
			logger.Error("Failed to create EC2 client", "error", err)
			os.Exit(1)
		}

		// Dynamically fetch AMIs (Amazon Linux 2 AMIs in us-east-2)
		amiFilter := &ec2.DescribeImagesInput{
			Owners: []string{"amazon"},
			Filters: []awstypes.Filter{
				{Name: aws.String("name"), Values: []string{"amzn2-ami-hvm-*-x86_64-gp2"}},
				{Name: aws.String("state"), Values: []string{"available"}},
			},
		}
		amiResult, err := ec2Client.Client().DescribeImages(ctx, amiFilter)
		if err != nil {
			logger.Error("Failed to fetch AMIs", "error", err)
			os.Exit(1)
		}
		var amiIDs []string
		for _, image := range amiResult.Images {
			if image.ImageId != nil {
				amiIDs = append(amiIDs, *image.ImageId)
			}
		}
		if len(amiIDs) == 0 {
			logger.Error("No AMIs found")
			os.Exit(1)
		}

		// Define instance types (safe to hardcode)
		instanceTypes := []string{
			"t2.micro",
			"t2.small",
		}

		// Dynamically fetch subnets
		subnetResult, err := ec2Client.Client().DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
		if err != nil {
			logger.Error("Failed to fetch subnets", "error", err)
			os.Exit(1)
		}
		var subnetIDs []string
		for _, subnet := range subnetResult.Subnets {
			if subnet.SubnetId != nil {
				subnetIDs = append(subnetIDs, *subnet.SubnetId)
			}
		}
		if len(subnetIDs) == 0 {
			logger.Error("No subnets found")
			os.Exit(1)
		}

		// Dynamically fetch key pairs
		keyResult, err := ec2Client.Client().DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{})
		if err != nil {
			logger.Error("Failed to fetch key pairs", "error", err)
			os.Exit(1)
		}
		keyNames := []string{""}
		for _, keyPair := range keyResult.KeyPairs {
			if keyPair.KeyName != nil {
				keyNames = append(keyNames, *keyPair.KeyName)
			}
		}

		// Randomly select parameters
		amiID := amiIDs[rand.Intn(len(amiIDs))]
		instanceType := instanceTypes[rand.Intn(len(instanceTypes))]
		subnetID := subnetIDs[rand.Intn(len(subnetIDs))]
		keyName := keyNames[rand.Intn(len(keyNames))]

		// Log the selected parameters for debugging
		logger.Info("Selected EC2 parameters",
			"ami_id", amiID,
			"instance_type", instanceType,
			"subnet_id", subnetID,
			"key_name", "test-value",
		)

		// Create the EC2 instance
		instanceID, err := createEC2Instance(ctx, ec2Client, amiID, instanceType, subnetID, keyName)
		if err != nil {
			logger.Error("Failed to create EC2 instance", "error", err)
			os.Exit(1)
		}

		logger.Info("EC2 instance created successfully", "instance_id", instanceID)

		// Wait for the instance to be running
		err = waitForInstanceRunning(ctx, ec2Client, instanceID)
		if err != nil {
			logger.Error("Failed to wait for instance to be running", "error", err)
			os.Exit(1)
		}

		logger.Info("EC2 instance is now running and ready for testing")
		fmt.Printf("To terminate the instance, run: go run main.go down %s\n", instanceID)

	case "down":
		// Create the EC2 client for setup/teardown
		ec2Client, err := customaws.NewEC2Client(ctx, false)
		if err != nil {
			logger.Error("Failed to create EC2 client", "error", err)
			os.Exit(1)
		}

		if len(os.Args) < 3 {
			fmt.Println("Please provide the instance ID to terminate.")
			fmt.Println("Example: go run main.go down i-1234567890abcdef0")
			os.Exit(1)
		}

		instanceID := os.Args[2]

		// Terminate the EC2 instance
		err = ec2Client.TerminateInstance(ctx, instanceID)
		if err != nil {
			logger.Error("Failed to terminate EC2 instance", "error", err)
			os.Exit(1)
		}

		logger.Info("EC2 instance terminated successfully", "instance_id", instanceID)

	case "detect":
		// Create the AWS client for drift detection
		awsClient, err := customaws.NewLiveAWSClient(ctx, logger)
		if err != nil {
			logger.Error("Failed to create AWS client", "error", err)
			os.Exit(1)
		}

		// Create the drift detector
		detector := usecases.NewDriftDetector(awsClient, logger)

		// Path to the Terraform state file
		tfStateFile := "terraform.tfstate"
		if len(os.Args) > 2 {
			tfStateFile = os.Args[2]
		}

		fmt.Println("Drift detection completed successfully")

		// Detect drift
		if err := detector.DetectDrift(tfStateFile); err != nil {
			logger.Error("Drift detection failed", "error", err)
			os.Exit(1)
		}

		fmt.Println("Drift detection completed successfully")

	default:
		fmt.Println("Invalid action. Use 'up' to create an instance, 'down' to terminate one, or 'detect' to detect drift.")
		os.Exit(1)
	}
}

// createEC2Instance creates a new EC2 instance and returns its instance ID.
func createEC2Instance(ctx context.Context, client customaws.EC2Client, amiID, instanceType, subnetID, keyName string) (string, error) {
	instanceID, err := client.CreateInstance(ctx, amiID, instanceType, subnetID, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}
	return instanceID, nil
}

// waitForInstanceRunning waits until the EC2 instance is in the "running" state.
func waitForInstanceRunning(ctx context.Context, client customaws.EC2Client, instanceID string) error {
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
