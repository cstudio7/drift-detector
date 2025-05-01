package aws

import (
	"context"
	"fmt"

	"github.com/cstudio7/drift-detector/internal/domain/entities"
	"github.com/cstudio7/drift-detector/internal/third_party/logger"
	"github.com/cstudio7/drift-detector/pkg/aws"
)

// LiveAWSClient is an implementation of AWSClient that interacts with live AWS services.
type LiveAWSClient struct {
	ec2Client EC2Client
	logger    logger.Logger
}

// NewLiveAWSClient creates a new LiveAWSClient.
func NewLiveAWSClient(ctx context.Context, logger logger.Logger) (*LiveAWSClient, error) {
	ec2Client, err := NewEC2Client(ctx, false) // useMock: false for live requests
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", entities.ErrFailedToCreateEC2Client)
	}

	return &LiveAWSClient{
		ec2Client: ec2Client,
		logger:    logger,
	}, nil
}

// FetchInstanceConfigs fetches the configurations of EC2 instances from AWS.
func (c *LiveAWSClient) FetchInstanceConfigs() ([]entities.InstanceConfig, error) {
	c.logger.Info("Fetching EC2 instance configurations from AWS")

	// Use the underlying ec2.Client directly to fetch all instances
	input := &aws.DescribeInstancesInput{}
	result, err := c.ec2Client.(*EC2ClientImpl).client.DescribeInstances(context.Background(), input)
	if err != nil {
		c.logger.Error("Failed to describe EC2 instances", "error", err)
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", entities.ErrFailedToFetchAWSConfigs)
	}

	var configs []entities.InstanceConfig
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			config := c.ec2Client.ToInstanceConfig(&instance)
			config.InstanceID = aws.ToString(instance.InstanceId) // Use aws.ToString from pkg/aws
			configs = append(configs, config)
			c.logger.Info("Fetched instance", "instance_id", config.InstanceID)
		}
	}

	c.logger.Info("Completed fetching EC2 instance configurations", "count", len(configs))
	return configs, nil
}
