# Contributing to Terraform Provider Astro

Welcome to the Terraform Provider Astro project! We're excited that you're interested in contributing. This document will guide you through the process of setting up your development environment, making changes, submitting pull requests, and reporting issues.

## Table of Contents

1. [Development Environment Setup](#development-environment-setup)
2. [Making Changes](#making-changes)
3. [Testing](#testing)
4. [Submitting Pull Requests](#submitting-pull-requests)
5. [Reporting Issues](#reporting-issues)
6. [Best Practices](#best-practices)
7. [Additional Resources](#additional-resources)

## Development Environment Setup

### Prerequisites

Ensure you have the following installed:
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.7
- [Go](https://golang.org/doc/install) >= 1.21

### Setting up the Provider for Local Development

1. Clone the repository:
   ```
   git clone https://github.com/astronomer/terraform-provider-astro.git
   cd terraform-provider-astro
   ```

2. Build the provider:
   ```
   make dep
   make build
   ```

3. Create a `.terraformrc` file in your home directory (`~`) with the following content:
   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/astronomer/astro" = "/path/to/your/terraform-provider-astro/bin"
     }
     direct {}
   }
   ```
   Replace `/path/to/your/terraform-provider-astro/bin` with the actual path to the `bin` directory in your cloned repository.

4. Create a `main.tf` file for testing:
   ```hcl
   terraform {
     required_providers {
       astro = {
         source = "astronomer/astro"
       }
     }
   }

   provider "astro" {
     organization_id = "<your-org-id>"
   }

   # Add resources and data sources here for testing
   ```

5. Set up your Astro API token:
   ```
   export ASTRO_API_TOKEN=<your-api-token>
   ```

## Making Changes

1. Create a new branch for your changes:
   ```
   git checkout -b feature/your-feature-name
   ```

2. Make your changes in the appropriate files. Common areas for changes include:
    - `internal/provider/` for provider logic
    - `internal/resources/` for resource implementations
    - `internal/datasources/` for data source implementations
    - `examples/` for example configurations

3. Update or add tests in the corresponding `*_test.go` files. If a new data source or resource is added, create a new test file in the `*_test.go` file for that feature.

4. Update documentation if your changes affect the provider's behavior or add new features.

## Testing

1. Run unit tests:
   ```
   make test
   ```

2. Run acceptance tests (these will create real resources in your Astro account):
   ```
   make testacc
   ```
   Note: Ensure all required environment variables are set as described in `internal/provider/provider_test_utils.go`.

3. Test your changes manually using the `main.tf` file you created earlier:
   ```
   terraform init
   terraform plan
   terraform apply
   ```

## Submitting Pull Requests

1. Commit your changes:
   ```
   git add .
   git commit -m "Description of your changes"
   ```

2. Push your branch to GitHub:
   ```
   git push origin feature/your-feature-name
   ```

3. Open a pull request on GitHub.

4. In your pull request description, include:
    - A clear description of the changes
    - Any related issue numbers
    - Steps to test the changes
    - Screenshots or code snippets if applicable

## Reporting Issues

If you encounter a bug or have a suggestion for improvement, please create an issue on the GitHub repository. This helps us track and address problems efficiently.

### Creating an Issue

1. Go to the [Issues page](https://github.com/astronomer/terraform-provider-astro/issues) of the repository.
2. Click on "New Issue".
3. Choose the appropriate issue template if available, or start with a blank issue.
4. Provide a clear and concise title that summarizes the issue.
5. In the description, include:
    - A detailed description of the bug or feature request
    - Steps to reproduce the issue (for bugs)
    - Expected behavior
    - Actual behavior (for bugs)
    - Screenshots or error messages, if applicable
    - Your environment details:
        - Terraform version
        - Terraform Provider Astro version
        - Operating System
        - Any other relevant information
6. Add appropriate labels to the issue (e.g., "bug", "enhancement", "documentation")
7. Submit the issue

### Best Practices for Issue Reporting

- Search existing issues before creating a new one to avoid duplicates.
- One issue per report. If you have multiple bugs or feature requests, create separate issues for each.
- Be responsive to questions or requests for additional information.
- If you find a solution to your problem before the issue is resolved, add a comment describing the solution for others who might encounter the same issue.

### Security Issues

If you discover a security vulnerability, please do NOT open an issue. Email security@astronomer.io instead.

## Best Practices

- Follow Go best practices and conventions.
- Ensure your code is well-commented and easy to understand.
- Keep your changes focused. If you're working on multiple features, submit separate pull requests.
- Update the README.md if your changes affect the overall usage of the provider.
- Add example configurations to the `examples/` directory for any new features or resources.
- If you're adding a new resource or data source, ensure you've added corresponding acceptance tests.

## Additional Resources

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [Go Documentation](https://golang.org/doc/)
- [Astro API Documentation](https://docs.astronomer.io/astro/api-overview)

Thank you for contributing to the Terraform Provider Astro project! Your efforts help improve the experience for all Astro users.