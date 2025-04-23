package aws

import (
    "context"
    "strings"
    "testing"

    awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"

    "github.com/cstudio7/drift-detector/internal/domain/entities"
)

func TestEC2ClientImpl_GetInstance(t *testing.T) {
    tests := []struct {
        name      string
        instanceID string
        expectErr bool
        expectedID string
    }{
        {
            name:       "ValidInstance",
            instanceID: "i-1234567890abcdef0",
            expectErr:  false,
            expectedID: "i-1234567890abcdef0",
        },
        {
            name:       "InstanceNotFound",
            instanceID: "i-nonexistent",
            expectErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, err := NewEC2Client(context.Background())
            if err != nil {
                t.Fatalf("Failed to create EC2 client: %v", err)
            }

            instance, err := client.GetInstance(context.Background(), tt.instanceID)
            if (err != nil) != tt.expectErr {
                t.Errorf("Expected error: %v, got error: %v", tt.expectErr, err)
            }
            if err != nil {
                if !strings.Contains(err.Error(), "instance not found") {
                    t.Errorf("Expected 'instance not found' error, got %v", err)
                }
                return
            }

            if *instance.InstanceId != tt.expectedID {
                t.Errorf("Expected instance ID %v, got %v", tt.expectedID, *instance.InstanceId)
            }
            if string(instance.InstanceType) != "t2.small" {
                t.Errorf("Expected instance type t2.small, got %v", instance.InstanceType)
            }
        })
    }
}

func TestEC2ClientImpl_ToInstanceConfig(t *testing.T) {
    instance := &awstypes.Instance{
        InstanceId:   stringPtr("i-1234567890abcdef0"),
        InstanceType: "t2.small",
        Tags: []awstypes.Tag{
            {Key: stringPtr("Name"), Value: stringPtr("prod-instance")},
        },
        SecurityGroups: []awstypes.GroupIdentifier{
            {GroupId: stringPtr("sg-12345678")},
        },
        SubnetId: stringPtr("subnet-12345678"),
        IamInstanceProfile: &awstypes.IamInstanceProfile{
            Arn: stringPtr("arn:aws:iam::123456789012:instance-profile/test-profile"),
        },
    }

    client, _ := NewEC2Client(context.Background())
    config := client.ToInstanceConfig(instance)

    expected := entities.InstanceConfig{
        InstanceType:      "t2.small",
        Tags:              map[string]string{"Name": "prod-instance"},
        SecurityGroupIDs:  []string{"sg-12345678"},
        SubnetID:          "subnet-12345678",
        IAMInstanceProfile: "arn:aws:iam::123456789012:instance-profile/test-profile",
    }

    if config.InstanceType != expected.InstanceType {
        t.Errorf("Expected InstanceType %q, got %q", expected.InstanceType, config.InstanceType)
    }
    if len(config.Tags) != len(expected.Tags) || config.Tags["Name"] != expected.Tags["Name"] {
        t.Errorf("Expected Tags %v, got %v", expected.Tags, config.Tags)
    }
    if len(config.SecurityGroupIDs) != len(expected.SecurityGroupIDs) || config.SecurityGroupIDs[0] != expected.SecurityGroupIDs[0] {
        t.Errorf("Expected SecurityGroupIDs %v, got %v", expected.SecurityGroupIDs, config.SecurityGroupIDs)
    }
    if config.SubnetID != expected.SubnetID {
        t.Errorf("Expected SubnetID %q, got %q", expected.SubnetID, config.SubnetID)
    }
    if config.IAMInstanceProfile != expected.IAMInstanceProfile {
        t.Errorf("Expected IAMInstanceProfile %q, got %q", expected.IAMInstanceProfile, config.IAMInstanceProfile)
    }
}

func stringPtr(s string) *string {
    return &s
}
