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

*   Output

*   Future Work and Improvements for Drift Detector 


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

# Drift Detector

Drift Detector is a tool that detects infrastructure drift between deployed AWS resources and Terraform state.

## Project Structure

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
├── pkg/
│   └── aws/
│       ├── aws.go                      # AWS SDK initialization logic and helpers
│       │             
│       └── ec2/
│           ├── ec2.go               # Type aliases and model wrappers for EC2 types
│                 
├── .env                              # AWS credentials (not committed)
├── terraform.tfstate                 # Terraform state file for drift detection
├── run-drift-detector.sh             # Script to automate the workflow
├── go.mod                            # Go module file
├── go.sum                            # Go dependencies
└── README.md                         # Documentation
```

Setup Instructions
------------------

1.  `git clone` cd drift-detector

2.  `go mod tidy`

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

4.  You should have the terraform state file(Json format), you can run with any of the file i gave provided which are in different format.
   NB: find terraform files at the root file. (The application can take different files)
5. We already have the script that should be able to run. So run  `chmod +x run-drift-detector.sh` to be able to run the script.    The Script is able to spin up some EC2 instances for test and also terminating them, but they are commented out. SO if you dont have a runnung EC2 instance, you could uncomment them for testing.


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

 `./run-drift-detector.sh`  

 this works provided 
- you have have your env set
- and you have a list of EC2 instances on your AWS, else you can uncomment in the  `./run-drift-detector.sh`  STEP 1 and 3 which only creates a list of instances in a situation where the user doesn't have any

*   **What the Script Does**:

   *   Runs the list command to show existing EC2 instances.

   *   Runs the detect command to check for drift (with up to 3 retries).




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



Result
--------------
Test
<img width="987" alt="Screenshot 2025-05-01 at 1 36 38 PM" src="https://github.com/user-attachments/assets/1b778075-0430-49e9-9ca6-79ef3c2e494e" />

Output
<img width="1335" alt="Screenshot 2025-05-01 at 1 37 42 PM" src="https://github.com/user-attachments/assets/71214eaa-9a4a-428e-8b91-e405ef15d038" />

<img width="1440" alt="Screenshot 2025-05-02 at 6 43 43 AM" src="https://github.com/user-attachments/assets/ead9a4bf-5b65-4b17-99a5-b4477e960632" />


Future Work and Improvements for Drift Detector
===============================================

This document outlines the current challenges faced by the drift-detector project and proposed improvements to enhance its functionality and robustness.

Challenges
----------

*   **Not having EC2 for testing**: Testing the drift-detector is constrained by the lack of access to live EC2 instances. Since the tool compares AWS EC2 configurations with Terraform state files, the absence of EC2 instances limits realistic testing scenarios. This affects the ability to:

    *   Validate drift detection against actual AWS API responses.

    *   Test handling of edge cases, such as missing or malformed instance data.

    *   Assess performance with large numbers of instances.

    *   Simulate transient AWS API errors or network issues.Without EC2 instances, testing relies on mocked data, which may not fully capture real-world complexities. This challenge necessitates alternative testing strategies, such as using AWS sandbox environments or localstack, which require additional setup.


Upcoming Works and Improvements
-------------------------------

*   **Reading multiple files leveraging Go dynamics**: Enhance the terraform.TFStateParser to support reading multiple Terraform state files concurrently. Currently, the detect command processes a single state file (e.g., terraform.tfstate). By leveraging Go's concurrency features, such as goroutines and channels, the tool can:

    *   Parse multiple state files in parallel, improving performance for large infrastructure setups.

    *   Aggregate configurations from multiple files (e.g., split by environment, module, or region) into a unified InstanceConfigSet.

    *   Support complex workflows where Terraform state is distributed across files.This improvement would make the drift-detector more scalable and flexible, enabling users to detect drift across an entire infrastructure in a single command. Example implementation outline:

    *   Modify ParseTFState to accept a slice of file paths.

    *   Use a sync.WaitGroup and goroutines to parse files concurrently.

    *   Collect results via a channel and merge into a single InstanceConfigSet.

    *   Handle errors gracefully, logging issues for specific files without failing the entire process.