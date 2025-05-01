package aws

import "github.com/aws/aws-sdk-go-v2/aws"

// Config is an alias for aws.Config from the AWS SDK.
type Config = aws.Config

// ToString is an alias for aws.ToString from the AWS SDK.
func ToString(s *string) string {
	return aws.ToString(s)
}
