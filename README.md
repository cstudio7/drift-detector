A Go application to detect configuration drift between AWS EC2 instances and Terraform state.

## Project Structure

- **cmd/drift-detector/**: Application entry point.
- **internal/domain/**: Core business logic and entities (Clean Architecture: Entities).
- **internal/usecases/**: Application-specific use cases (Clean Architecture: Use Cases).
- **internal/interfaces/**: Interface adapters for AWS, Terraform, and logging (Clean Architecture: Interface Adapters).
- **internal/test/**: Unit tests for all layers.

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/cstudio7/drift-detector.git
   cd drift-detector