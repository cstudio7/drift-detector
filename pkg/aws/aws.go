package aws

import "github.com/aws/aws-sdk-go-v2/aws"

// Config is an alias for aws.Config from the AWS SDK.
type Config = aws.Config

// ToString is an alias for aws.ToString from the AWS SDK.
func ToString(s *string) string {
	return aws.ToString(s)
}

// String creates a pointer to a string.
func String(s string) *string {
	return aws.String(s)
}

// Int32 creates a pointer to an int32.
func Int32(i int32) *int32 {
	return aws.Int32(i)
}
