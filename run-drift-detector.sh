#!/bin/bash

# Ensure the script exits on any error
set -e

# Check for required files
if [ ! -f ".env" ]; then
    echo "Error: .env file not found in the current directory."
    echo "Please create a .env file with AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)."
    exit 1
fi

if [ ! -f "terraform.tfstate" ]; then
    echo "Error: terraform.tfstate file not found in the current directory."
    echo "Please create a terraform.tfstate file for drift detection."
    exit 1
fi

# Initialize INSTANCE_ID variable
INSTANCE_ID=""

# Trap to clean up on script failure
cleanup() {
    if [ -n "$INSTANCE_ID" ]; then
        echo "Cleaning up instance $INSTANCE_ID..."
        go run cmd/drift-detector/main.go down "$INSTANCE_ID" || true
    fi
}
trap cleanup EXIT

# Step 1: Create an EC2 instance (up)
echo "Creating EC2 instance..."
if ! go run cmd/drift-detector/main.go up 2>&1 | tee /dev/tty; then
    echo "Error: 'up' command failed."
    exit 1
fi

# Capture the output after running the command
UP_OUTPUT=$(go run cmd/drift-detector/main.go up 2>&1)

# Extract the instance ID from the output
INSTANCE_ID=$(echo "$UP_OUTPUT" | grep "To terminate the instance" | grep -o "i-[a-z0-9]\+")
if [ -z "$INSTANCE_ID" ]; then
    echo "Error: Failed to extract instance ID from the 'up' command output."
    echo "Output: $UP_OUTPUT"
    exit 1
fi
echo "Extracted instance ID: $INSTANCE_ID"

# Wait for the instance to be fully running
echo "Waiting for instance to be fully running..."
sleep 30

# Step 2: Detect drift (detect) with retries
echo "Detecting drift..."
for i in {1..3}; do
    if go run cmd/drift-detector/main.go detect terraform.tfstate; then
        break
    else
        echo "Drift detection failed, retrying ($i/3)..."
        sleep 10
        if [ "$i" -eq 3 ]; then
            echo "Error: Drift detection failed after 3 attempts."
            exit 1
        fi
    fi
done

# Step 3: Terminate the EC2 instance (down)
echo "Terminating EC2 instance ($INSTANCE_ID)..."
go run cmd/drift-detector/main.go down "$INSTANCE_ID"

echo "Application run completed successfully."