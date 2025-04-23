package aws

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "runtime"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"

    "github.com/cstudio7/drift-detector/internal/domain/entities"
)

// getTestDataPath resolves the path to the testdata directory dynamically.
func getTestDataPath(filename string) (string, error) {
    _, callerFile, _, ok := runtime.Caller(0)
    if !ok {
        return "", fmt.Errorf("failed to get caller information")
    }
    // Navigate to the testdata directory at the project root
    // internal/interfaces/aws/ec2.go -> testdata/
    baseDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(callerFile)))) // Up four levels to project root
    testDataPath := filepath.Join(baseDir, "testdata", filename)
    return testDataPath, nil
}

// EC2Client defines the interface for interacting with AWS EC2.
type EC2Client interface {
    GetInstance(ctx context.Context, instanceID string) (*awstypes.Instance, error)
    ToInstanceConfig(instance *awstypes.Instance) entities.InstanceConfig
}

// EC2ClientImpl is the implementation of EC2Client.
type EC2ClientImpl struct {
    client *ec2.Client
}

// NewEC2Client creates a new EC2ClientImpl.
func NewEC2Client(ctx context.Context) (*EC2ClientImpl, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    return &EC2ClientImpl{client: ec2.NewFromConfig(cfg)}, nil
}

// GetInstance retrieves an EC2 instance by ID (mocked for testing).
func (c *EC2ClientImpl) GetInstance(ctx context.Context, instanceID string) (*awstypes.Instance, error) {
    filePath, err := getTestDataPath("sample-ec2.json")
    if err != nil {
        return nil, fmt.Errorf("failed to resolve test data path: %w", err)
    }

    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read mock data: %w", err)
    }

    var resp struct {
        Reservations []struct {
            Instances []awstypes.Instance
        }
    }
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, fmt.Errorf("failed to parse mock data: %w", err)
    }

    if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
        return nil, entities.ErrInstanceNotFound
    }

    if resp.Reservations[0].Instances[0].InstanceId == nil || *resp.Reservations[0].Instances[0].InstanceId != instanceID {
        return nil, entities.ErrInstanceNotFound
    }

    return &resp.Reservations[0].Instances[0], nil
}

// ToInstanceConfig converts an EC2 instance to an InstanceConfig struct.
func (c *EC2ClientImpl) ToInstanceConfig(instance *awstypes.Instance) entities.InstanceConfig {
    config := entities.InstanceConfig{
        InstanceType:      string(instance.InstanceType),
        Tags:              make(map[string]string),
        SecurityGroupIDs:  make([]string, 0),
        SubnetID:          "",
        IAMInstanceProfile: "",
    }
    for _, tag := range instance.Tags {
        if tag.Key != nil && tag.Value != nil {
            config.Tags[*tag.Key] = *tag.Value
        }
    }
    for _, sg := range instance.SecurityGroups {
        if sg.GroupId != nil {
            config.SecurityGroupIDs = append(config.SecurityGroupIDs, *sg.GroupId)
        }
    }
    if instance.SubnetId != nil {
        config.SubnetID = *instance.SubnetId
    }
    if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
        config.IAMInstanceProfile = *instance.IamInstanceProfile.Arn
    }
    return config
}
