package entities

import "errors"

var (
	// ErrInstanceNotFound is returned when an instance is not found in AWS or Terraform state.
	ErrInstanceNotFound = errors.New("instance not found")

	// ErrFetchAWSConfigs is returned when fetching AWS instance configurations fails.
	ErrFetchAWSConfigs = errors.New("failed to fetch AWS configurations")

	// ErrInvalidTerraformState is returned when the Terraform state file is invalid or cannot be parsed.
	ErrInvalidTerraformState = errors.New("invalid Terraform state")

	// ErrConfigComparison is returned when comparing AWS and Terraform configurations fails due to invalid data.
	ErrConfigComparison = errors.New("failed to compare configurations")

	// ErrEmptyConfigs is returned when no configurations are found in AWS or Terraform state.
	ErrEmptyConfigs = errors.New("no configurations found")

	// ErrFailedToCreateEC2Client is returned when creating an EC2 client fails.
	ErrFailedToCreateEC2Client = errors.New("failed to create EC2 client")

	// ErrFailedToFetchAMIs is returned when fetching AMIs from AWS fails.
	ErrFailedToFetchAMIs = errors.New("failed to fetch AMIs")

	// ErrNoAMIsFound is returned when no AMIs are found.
	ErrNoAMIsFound = errors.New("no AMIs found")

	// ErrFailedToFetchSubnets is returned when fetching subnets from AWS fails.
	ErrFailedToFetchSubnets = errors.New("failed to fetch subnets")

	// ErrNoSubnetsFound is returned when no subnets are found.
	ErrNoSubnetsFound = errors.New("no subnets found")

	// ErrFailedToFetchKeyPairs is returned when fetching key pairs from AWS fails.
	ErrFailedToFetchKeyPairs = errors.New("failed to fetch key pairs")

	// ErrFailedToCreateInstance is returned when creating an EC2 instance fails.
	ErrFailedToCreateInstance = errors.New("failed to create EC2 instance")

	// ErrFailedToTerminateInstance is returned when terminating an EC2 instance fails.
	ErrFailedToTerminateInstance = errors.New("failed to terminate EC2 instance")

	// ErrFailedToWaitForInstance is returned when waiting for an EC2 instance to become running fails.
	ErrFailedToWaitForInstance = errors.New("failed to wait for instance to be running")

	// ErrInvalidAction is returned when an invalid action is provided in the CLI.
	ErrInvalidAction = errors.New("invalid action")

	// ErrFailedToFetchAWSConfigs indicates failure in retrieving AWS configuration (e.g., credentials or region).
	ErrFailedToFetchAWSConfigs = errors.New("failed to fetch AWS configuration")

	// ErrMissingInstanceID indicates that the instance ID was not provided when required.
	ErrMissingInstanceID = errors.New("please provide the instance ID to terminate")

	// ErrDriftDetectionFailed indicates a failure during drift detection.
	ErrDriftDetectionFailed = errors.New("drift detection failed")
)
