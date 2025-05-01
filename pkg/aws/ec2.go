package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client is an alias for ec2.Client from the AWS SDK.
type EC2Client = ec2.Client

// DescribeInstancesInput is an alias for ec2.DescribeInstancesInput.
type DescribeInstancesInput = ec2.DescribeInstancesInput

// DescribeInstancesOutput is an alias for ec2.DescribeInstancesOutput.
type DescribeInstancesOutput = ec2.DescribeInstancesOutput

// Instance is an alias for types.Instance from the AWS SDK.
type Instance = types.Instance

// RunInstancesInput is an alias for ec2.RunInstancesInput.
type RunInstancesInput = ec2.RunInstancesInput

// TerminateInstancesInput is an alias for ec2.TerminateInstancesInput.
type TerminateInstancesInput = ec2.TerminateInstancesInput

// InstanceType is an alias for types.InstanceType.
type InstanceType = types.InstanceType

// TagSpecification is an alias for types.TagSpecification.
type TagSpecification = types.TagSpecification

// Tag is an alias for types.Tag.
type Tag = types.Tag

// ResourceType is an alias for types.ResourceType.
type ResourceType = types.ResourceType
