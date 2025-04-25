Drift Detector
==============

Drift Detector is a Go application designed to detect configuration drift between existing AWS EC2 instances and a Terraform state file. It allows you to:

*   List existing EC2 instances (optional - We included code for creating instance for testing).

*   Detect drift by comparing the live AWS state with a Terraform state file (detect command).


The application includes a script to automate the workflow, making it easy to list instances and detect drift in a single command.

Table of Contents
-----------------

*   Prerequisites

*   Project Structure

*   Setup Instructions

*   Running the Application

*   Testing the Application

*   Troubleshooting

*   Best Practices


Prerequisites
-------------

Before running the Drift Detector, ensure you have the following:

### Go

*   Install Go (version 1.20 or later): [Download Go](https://golang.org/dl/)

*   go version


### AWS Account and Credentials

*   An AWS account with permissions for:

   *   ec2:DescribeInstances

*   AWS credentials configured in a .env file (see Setup Instructions).


### Terraform State File

*   A terraform.tfstate file in the root directory for drift detection. An example is provided in the setup instructions, but we use testdata/sample-tfstate.json for this test


### Bash Shell (for running the script)

*   The run-drift-detector.sh script requires a Bash-compatible shell (e.g., on macOS, Linux, or WSL on Windows).


Project Structure
-----------------

### The project is organized as follows:

```text
drift-detector/
├── cmd/
│   └── drift-detector/
│       └── main.go               # Main application entry point
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   │   └── instance_config.go    # InstanceConfig type definition
│   │   └── services/                 # Business logic for drift detection (not modified)
│   ├── interfaces/
│   │   ├── aws/
│   │   │   ├── client.go             # AWSClient interface and LiveAWSClient implementation
│   │   │   ├── client_test.go        # Tests for LiveAWSClient
│   │   │   └── ec2.go                # EC2Client interface and implementation
│   │   ├── logger/                   # Logger interface (not modified)
│   │   └── terraform/                # Terraform interface (not modified)
│   └── usecases/                     # Drift detection logic (not modified)
├── .env                              # AWS credentials (not committed)
├── terraform.tfstate                 # Terraform state file for drift detection
├── run-drift-detector.sh             # Script to automate the workflow
├── go.mod                            # Go module file
├── go.sum                            # Go dependencies
└── README.md                         # Documentation
```

Setup Instructions
------------------

1.  git clone cd drift-detector

2.  go mod tidy

3.  AWS_ACCESS_KEY_ID=your-access-key-id
    AWS_SECRET_ACCESS_KEY=your-secret-access-key   
    AWS_REGION=us-east-1  
    SUBNET_IDS=subnet-0000000000,subnet-00000000

   *   Replace your-access-key-id and your-secret-access-key with your AWS credentials.

   *   Ensure AWS\_REGION matches the region of your subnets (e.g., us-east-2).

   *   **Note**: Do not commit the .env file to version control. Add it to .gitignore.
       AWS_ACCESS_KEY_ID=
       AWS_SECRET_ACCESS_KEY=
       AWS_REGION=us-east-1
       SUBNET_IDS=
4.  { "Version": "2012-10-17", "Statement": \[ { "Effect": "Allow", "Action": \[ "ec2:DescribeInstances" \], "Resource": "\*" } \]}Test your credentials:aws sts get-caller-identity --region us-east-2

5.  { "version": 4, "terraform\_version": "1.0.0", "serial": 1, "lineage": "fake-lineage", "outputs": {}, "resources": \[ { "mode": "managed", "type": "aws\_instance", "name": "example", "provider": "provider\[\\"registry.terraform.io/hashicorp/aws\\"\]", "instances": \[ { "attributes": { "id": "i-0987654321fedcba0", "instance\_type": "t2.micro", "tags": { "Name": "example-instance" }, "subnet\_id": "subnet-0fd29464681088be4", "security\_groups": \["sg-0a1b2c3d4e5f6g7h8"\], "iam\_instance\_profile": "" } } \] } \]}

   *   aws ec2 describe-instances --region us-east-2 --query 'Reservations\[\*\].Instances\[\*\].InstanceId' --output table

   *   aws ec2 describe-subnets --region us-east-2 --query 'Subnets\[\*\].SubnetId' --output table

   *   aws ec2 describe-security-groups --region us-east-2 --query 'SecurityGroups\[\*\].GroupId' --output table

6.  chmod +x run-drift-detector.sh


Running the Application
-----------------------

The application can be run in two ways: manually (using individual commands) or automatically (using the script).

### Option 1: Run Manually

1.  go run cmd/drift-detector/main.go list (list can be up or down or detect)

   *   This lists the instance IDs of all non-terminated EC2 instances in your AWS account.

   *   2025/04/24 00:00:00 \[INFO\] Found EC2 instances count=2Existing EC2 instances:- i-0abcdef1234567890- i-1bcdef234567890a

   *   2025/04/24 00:00:00 \[INFO\] No EC2 instances foundNo EC2 instances found in the account.

2.  go run cmd/drift-detector/main.go detect terraform.tfstate

   *   This compares the live AWS state (all non-terminated instances) with the terraform.tfstate file and reports any drift.

   *   2025/04/24 00:00:00 \[INFO\] Fetching EC2 instance configurations from AWS2025/04/24 00:00:00 \[INFO\] Fetched instance instance\_id=i-0abcdef12345678902025/04/24 00:00:00 \[INFO\] Fetched instance instance\_id=i-1bcdef234567890a2025/04/24 00:00:00 \[INFO\] Completed fetching EC2 instance configurations count=22025/04/24 00:00:00 \[INFO\] Drift detected: Instance not found in Terraform state instance\_id=i-0abcdef12345678902025/04/24 00:00:00 \[INFO\] Drift detected: Instance not found in Terraform state instance\_id=i-1bcdef234567890aDrift detection completed successfully


### Option 2: Run Automatically with the Script (Recommended)

The run-drift-detector.sh script automates the workflow:   

 `  ./run-drift-detector.sh `  

*   **What the Script Does**:

   *   Runs the list command to show existing EC2 instances.

   *   Runs the detect command to check for drift (with up to 3 retries).

*   Listing existing EC2 instances...2025/04/24 00:00:00 \[INFO\] Found EC2 instances count=2Existing EC2 instances:- i-0abcdef1234567890- i-1bcdef234567890aDetecting drift...2025/04/24 00:00:00 \[INFO\] Fetching EC2 instance configurations from AWS2025/04/24 00:00:00 \[INFO\] Fetched instance instance\_id=i-0abcdef12345678902025/04/24 00:00:00 \[INFO\] Fetched instance instance\_id=i-1bcdef234567890a2025/04/24 00:00:00 \[INFO\] Completed fetching EC2 instance configurations count=22025/04/24 00:00:00 \[INFO\] Drift detected: Instance not found in Terraform state instance\_id=i-0abcdef12345678902025/04/24 00:00:00 \[INFO\] Drift detected: Instance not found in Terraform state instance\_id=i-1bcdef234567890aDrift detection completed successfullyApplication run completed successfully.


Testing the Application
-----------------------

Testing ensures that the Drift Detector application works as expected. Below are steps for both functional testing (using the application) and unit testing.

To run all test use: `go test -v ./...`

### Unit Testing

The codebase includes unit tests for several packages. To run the tests and generate a coverage report:

` go test -coverprofile=cover.out ./internal/...`

The current test coverage is:

*   internal/domain/services: 90.0% of statements

*   internal/interfaces/aws: 96.0% of statements

*   internal/interfaces/logger: 100.0% of statements

*   internal/interfaces/terraform: 90.0% of statements

*   internal/usecases: 90.0% of statements

To add tests for the internal/interfaces/aws package, you can create a mock EC2Client and test the LiveAWSClient logic. Here’s an example:

1.  **Create a Test File**:

   *   Create a file named client\_test.go in the internal/interfaces/aws directory.

   *   package awsimport ( "errors" "testing" "github.com/aws/aws-sdk-go-v2/service/ec2" "github.com/aws/aws-sdk-go-v2/service/ec2/types" "github.com/cstudio7/drift-detector/internal/domain/entities" "github.com/cstudio7/drift-detector/internal/interfaces/logger")type mockEC2Client struct { instances \[\]types.Instance err error}func (m \*mockEC2Client) Client() \*ec2.Client { return &ec2.Client{}}func (m \*mockEC2Client) GetInstance(ctx context.Context, instanceID string) (\*types.Instance, error) { for \_, instance := range m.instances { if \*instance.InstanceId == instanceID { return &instance, nil } } return nil, fmt.Errorf("instance %s not found", instanceID)}func TestLiveAWSClient\_FetchInstanceConfigs(t \*testing.T) { logger := logger.NewStdLogger() // Mock EC2 client with some instances instanceID := "i-1234567890abcdef0" instanceType := types.InstanceTypeT2Micro tags := \[\]types.Tag{ {Key: aws.String("Name"), Value: aws.String("test-instance")}, } mockClient := &mockEC2Client{ instances: \[\]types.Instance{ { InstanceId: &instanceID, InstanceType: instanceType, State: &types.InstanceState{Name: "running"}, Tags: tags, }, }, } // Create LiveAWSClient with the mock EC2 client client := &LiveAWSClient{ ec2Client: mockClient, logger: logger, } // Override the instances for the test client.ec2Client.(\*mockEC2Client).instances = \[\]types.Instance{ { InstanceId: &instanceID, InstanceType: instanceType, State: &types.InstanceState{Name: "running"}, Tags: tags, }, } // Fetch instance configs configs, err := client.FetchInstanceConfigs() if err != nil { t.Fatalf("FetchInstanceConfigs failed: %v", err) } // Verify the results if len(configs) != 1 { t.Fatalf("Expected 1 instance config, got %d", len(configs)) } config := configs\[0\] if config.InstanceID != instanceID { t.Errorf("Expected InstanceID %s, got %s", instanceID, config.InstanceID) } if config.InstanceType != string(instanceType) { t.Errorf("Expected InstanceType %s, got %s", string(instanceType), config.InstanceType) } if config.Tags\["Name"\] != "test-instance" { t.Errorf("Expected tag Name=test-instance, got %s", config.Tags\["Name"\]) }}func TestLiveAWSClient\_FetchInstanceConfigs\_Error(t \*testing.T) { logger := logger.NewStdLogger() // Mock EC2 client with an error mockClient := &mockEC2Client{ err: errors.New("AWS API error"), } // Create LiveAWSClient with the mock EC2 client client := &LiveAWSClient{ ec2Client: mockClient, logger: logger, } // Override the instances to trigger an error client.ec2Client.(\*mockEC2Client).instances = nil // Fetch instance configs \_, err := client.FetchInstanceConfigs() if err == nil { t.Fatal("Expected FetchInstanceConfigs to fail, but it succeeded") } if err.Error() != "failed to describe instances: AWS API error" { t.Errorf("Expected error 'failed to describe instances: AWS API error', got '%v'", err) }}

2.  go test ./internal/interfaces/aws -v

3.  **Expected Result**:

   *   The tests pass, verifying that FetchInstanceConfigs correctly fetches and converts EC2 instance configurations, including error handling.


Troubleshooting
---------------

If the application or script fails, here are common issues and solutions:

1.  **Script Fails at "Listing existing EC2 instances..."**:

   *   **Cause**: The list command might be failing due to AWS API errors or network issues.

   *   go run cmd/drift-detector/main.go list

   *   Check for errors like Failed to create EC2 client.

2.  **AWS API Errors (e.g., AccessDenied)**:

   *   **Cause**: Missing permissions for DescribeInstances.

   *   **Fix**: Update your IAM policy as shown in the setup instructions.

3.  **Invalid AWS Credentials**:

   *   **Cause**: Expired or incorrect credentials in .env.

   *   aws sts get-caller-identity --region us-east-2

   *   Update .env with valid credentials.

4.  **No Instances Found**:

   *   **Cause**: There are no non-terminated EC2 instances in your AWS account.

   *   **Fix**: Create an instance manually in the AWS Console or using the AWS CLI (see Test 1 in the Testing section).

5.  **Drift Detection Fails**:

   *   **Cause**: The instances in AWS don’t match the terraform.tfstate file.

   *   **Fix**: Update terraform.tfstate to match an existing instance, or create an instance that matches the state file.

6.  **Network Issues**:

   *   **Cause**: Unstable network causing AWS API calls to fail.

   *   aws ec2 describe-instances --region us-east-2

   *   Retry on a stable connection.


Best Practices
--------------

1.  **Avoid Hardcoding Sensitive Data**:

   *   The application dynamically fetches instance data to avoid hardcoding. If you need to hardcode values (e.g., for testing), store them in environment variables or a configuration file, and do not commit them to version control.

2.  **Secure AWS Credentials**:

   *   Keep the .env file out of version control by adding it to .gitignore.

   *   Use IAM roles or temporary credentials for production environments.

3.  **Monitor AWS Resources**:

   *   Since this application only lists instances, it doesn’t create or terminate resources. However, always check the AWS Management Console (EC2 > Instances) to manage your instances.

4.  **Test in a Sandbox Environment**:

   *   Run this application in a non-production AWS account to avoid accidental changes to critical infrastructure.

5.  **Regularly Review Instances**:

   *   Periodically check your AWS account for unused or unnecessary EC2 instances to avoid unexpected costs.
