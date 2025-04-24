package aws

import "github.com/cstudio7/drift-detector/internal/domain/entities"

// AWSClient defines the interface for interacting with AWS services.
type AWSClient interface {
	FetchInstanceConfigs() ([]entities.InstanceConfig, error)
}
