package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client is an alias for ec2.Client from the AWS SDK.
type EC2Client = ec2.Client

// DescribeInstancesInput is an alias for ec2.DescribeInstancesInput.
type DescribeInstancesInput = ec2.DescribeInstancesInput

// DescribeInstancesOutput is an alias for ec2.DescribeInstancesOutput.
type DescribeInstancesOutput = ec2.DescribeInstancesOutput

// Instance is an alias for types.Instance from the AWS SDK.
type Instance = ec2types.Instance

// RunInstancesInput is an alias for ec2.RunInstancesInput.
type RunInstancesInput = ec2.RunInstancesInput

// TerminateInstancesInput is an alias for ec2.TerminateInstancesInput.
type TerminateInstancesInput = ec2.TerminateInstancesInput

// InstanceType is an alias for types.InstanceType.
type InstanceType = ec2types.InstanceType

// TagSpecification is an alias for types.TagSpecification.
type TagSpecification = ec2types.TagSpecification

// Tag is an alias for types.Tag.
type Tag = ec2types.Tag

// ResourceType is an alias for types.ResourceType.
type ResourceType = ec2types.ResourceType

// ResourceTypeInstance is a constant for the instance resource type.
const ResourceTypeInstance = ec2types.ResourceTypeInstance

// Initialize sets up and returns a new EC2 client.
func Initialize(ctx context.Context) (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

// DescribeImagesInput is an alias for ec2.DescribeImagesInput.
type DescribeImagesInput = ec2.DescribeImagesInput

// DescribeImagesOutput is an alias for ec2.DescribeImagesOutput.
type DescribeImagesOutput = ec2.DescribeImagesOutput

// Image is an alias for ec2types.Image.
type Image = ec2types.Image

// DescribeSubnetsInput is an alias for ec2.DescribeSubnetsInput.
type DescribeSubnetsInput = ec2.DescribeSubnetsInput

// DescribeSubnetsOutput is an alias for ec2.DescribeSubnetsOutput.
type DescribeSubnetsOutput = ec2.DescribeSubnetsOutput

// Subnet is an alias for ec2types.Subnet.
type Subnet = ec2types.Subnet

// DescribeKeyPairsInput is an alias for ec2.DescribeKeyPairsInput.
type DescribeKeyPairsInput = ec2.DescribeKeyPairsInput

// DescribeKeyPairsOutput is an alias for ec2.DescribeKeyPairsOutput.
type DescribeKeyPairsOutput = ec2.DescribeKeyPairsOutput

// KeyPairInfo is an alias for ec2types.KeyPairInfo.
type KeyPairInfo = ec2types.KeyPairInfo

// Filter is an alias for ec2types.Filter.
type Filter = ec2types.Filter

// InstanceState is an alias for ec2types.InstanceState.
type InstanceState = ec2types.InstanceState
