package entities

import "errors"

var (
	// ErrInstanceNotFound is returned when an instance is not found in AWS or Terraform state.
	ErrInstanceNotFound = errors.New("instance not found")

	//// ErrFetchAWSConfigs is returned when fetching AWS instance configurations fails.
	//ErrFetchAWSConfigs = errors.New("failed to fetch AWS configurations")
	//
	//// ErrInvalidTerraformState is returned when the Terraform state file is invalid or cannot be parsed.
	//ErrInvalidTerraformState = errors.New("invalid Terraform state")
	//
	//// ErrMissingResource is returned when an expected resource (e.g., aws_instance) is missing in the Terraform state.
	//ErrMissingResource = errors.New("missing resource in Terraform state")
	//
	//// ErrConfigComparison is returned when comparing AWS and Terraform configurations fails due to invalid data.
	//ErrConfigComparison = errors.New("failed to compare configurations")
	//
	//// ErrEmptyConfigs is returned when no configurations are found in AWS or Terraform state.
	//ErrEmptyConfigs = errors.New("no configurations found")
)
