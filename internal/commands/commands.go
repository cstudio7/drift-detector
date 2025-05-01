package commands

import (
	"context"
	"fmt"
	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	customaws "github.com/cstudio7/drift-detector/internal/third_party/aws"
	"github.com/cstudio7/drift-detector/internal/third_party/logger"
	"github.com/cstudio7/drift-detector/internal/usecases"
	"math/rand/v2"
)

// DriftCommand handles the CLI commands for the drift detector.
type DriftCommand struct {
	logger    logger.Logger
	awsClient customaws.AWSClient     // For drift detection
	ec2Client customaws.EC2Client     // For setup/teardown
	detector  *usecases.DriftDetector // Review interface
	ctx       context.Context
}

// NewDriftCommand creates a new DriftCommand with the provided logger.
func NewDriftCommand(logger logger.Logger) *DriftCommand {
	ctx := context.Background()
	return &DriftCommand{
		logger: logger,
		ctx:    ctx,
	}
}

// Run executes the specified command with the given arguments.
func (c *DriftCommand) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no action provided")
	}

	action := args[0]

	switch action {
	case "up":
		// Initialize EC2 client for setup/teardown
		ec2Client, err := customaws.NewEC2Client(c.ctx, false)
		if err != nil {
			return fmt.Errorf("failed to create EC2 client: %w", err)
		}
		c.ec2Client = ec2Client

		// Dynamically fetch AMIs
		amiFilter := &ec2.DescribeImagesInput{
			Owners: []string{"amazon"},
			Filters: []awstypes.Filter{
				{Name: aws.String("name"), Values: []string{"amzn2-ami-hvm-*-x86_64-gp2"}},
				{Name: aws.String("state"), Values: []string{"available"}},
			},
		}
		amiResult, err := ec2Client.Client().DescribeImages(c.ctx, amiFilter)
		if err != nil {
			return fmt.Errorf("failed to fetch AMIs: %w", entities.ErrFailedToFetchAMIs)
		}
		var amiIDs []string
		for _, image := range amiResult.Images {
			if image.ImageId != nil {
				amiIDs = append(amiIDs, *image.ImageId)
			}
		}
		if len(amiIDs) == 0 {
			return fmt.Errorf("no AMIs found: %w", entities.ErrNoAMIsFound)
		}

		// Define instance types
		instanceTypes := []string{
			"t2.micro",
			"t2.small",
		}

		// Dynamically fetch subnets
		subnetResult, err := ec2Client.Client().DescribeSubnets(c.ctx, &ec2.DescribeSubnetsInput{})
		if err != nil {
			return fmt.Errorf("failed to fetch subnets: %w", entities.ErrFailedToFetchSubnets)
		}
		var subnetIDs []string
		for _, subnet := range subnetResult.Subnets {
			if subnet.SubnetId != nil {
				subnetIDs = append(subnetIDs, *subnet.SubnetId)
			}
		}
		if len(subnetIDs) == 0 {
			return fmt.Errorf("failed to fetch key pairs: %w", entities.ErrNoSubnetsFound)
		}

		// Dynamically fetch key pairs
		keyResult, err := ec2Client.Client().DescribeKeyPairs(c.ctx, &ec2.DescribeKeyPairsInput{})
		if err != nil {
			return fmt.Errorf("failed to fetch key pairs: %w", entities.ErrFailedToFetchKeyPairs)
		}
		keyNames := []string{""}
		for _, keyPair := range keyResult.KeyPairs {
			if keyPair.KeyName != nil {
				keyNames = append(keyNames, *keyPair.KeyName)
			}
		}

		// Randomly select parameters using rand/v2
		amiID := amiIDs[rand.IntN(len(amiIDs))]
		instanceType := instanceTypes[rand.IntN(len(instanceTypes))]
		subnetID := subnetIDs[rand.IntN(len(subnetIDs))]
		keyName := keyNames[rand.IntN(len(keyNames))]

		// Log selected parameters
		c.logger.Info("Selected EC2 parameters",
			"ami_id", amiID,
			"instance_type", instanceType,
			"subnet_id", subnetID,
			"key_name", "test-value",
		)

		// Create the EC2 instance
		instanceID, err := c.createEC2Instance(amiID, instanceType, subnetID, keyName)
		if err != nil {
			return fmt.Errorf("failed to create EC2 client: %w", entities.ErrFailedToCreateEC2Client)
		}

		c.logger.Info("EC2 instance created successfully", "instance_id", instanceID)

		// Wait for the instance to be running
		err = c.waitForInstanceRunning(instanceID)
		if err != nil {
			return fmt.Errorf("timeout waiting for instance %s to be running: %w", instanceID, entities.ErrFailedToWaitForInstance)
		}

		c.logger.Info("EC2 instance is now running and ready for testing")
		fmt.Printf("To terminate the instance, run: go run main.go down %s\n", instanceID)

	case "down":
		if len(args) < 2 {
			return entities.ErrMissingInstanceID
		}
		instanceID := args[1]

		// Initialize EC2 client for setup/teardown
		ec2Client, err := customaws.NewEC2Client(c.ctx, false)
		if err != nil {
			return fmt.Errorf("failed to create instance: %w", entities.ErrFailedToCreateInstance)
		}
		c.ec2Client = ec2Client

		// Terminate the EC2 instance
		err = ec2Client.TerminateInstance(c.ctx, instanceID)
		if err != nil {
			return fmt.Errorf("failed to terminate EC2 instance: %w", entities.ErrFailedToTerminateInstance)

		}

		c.logger.Info("EC2 instance terminated successfully", "instance_id", instanceID)

	case "detect":
		// Use a default Terraform state file if none is provided
		tfStateFile := "terraform.tfstate" // Default file
		if len(args) >= 2 {
			tfStateFile = args[1]
		}

		// Initialize AWS client for drift detection
		awsClient, err := customaws.NewLiveAWSClient(c.ctx, c.logger)
		if err != nil {
			return fmt.Errorf("failed to create AWS Instance: %w", entities.ErrFailedToCreateInstance)
		}
		c.awsClient = awsClient

		// Create the drift detector
		c.detector = usecases.NewDriftDetector(awsClient, c.logger)

		// Perform drift detection
		err = c.detector.DetectDrift(tfStateFile)
		if err != nil {
			return fmt.Errorf("%w: %v", entities.ErrDriftDetectionFailed, err)
		}

		c.logger.Info("Drift detection completed successfully")

	default:
		return fmt.Errorf("invalid action: %s. Use 'up', 'down', or 'detect'", action)
	}

	return nil
}

// createEC2Instance creates a new EC2 instance and returns its instance ID.
func (c *DriftCommand) createEC2Instance(amiID, instanceType, subnetID, keyName string) (string, error) {
	instanceID, err := c.ec2Client.CreateInstance(c.ctx, amiID, instanceType, subnetID, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", entities.ErrFailedToCreateInstance)
	}
	return instanceID, nil
}

// waitForInstanceRunning waits until the EC2 instance is in the "running" state.
func (c *DriftCommand) waitForInstanceRunning(instanceID string) error {
	for i := 0; i < 30; i++ { // Retry up to 30 times (5 minutes total)
		instance, err := c.ec2Client.GetInstance(c.ctx, instanceID)
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

	return fmt.Errorf("timeout waiting for instance %s to be running: %w", instanceID, entities.ErrFailedToWaitForInstance)

}
