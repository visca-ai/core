# VISCA

**Open Source RL Environment Collaboration Platform**

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)
[![Build Status](https://img.shields.io/github/workflow/status/visca-ai/visca/CI)](https://github.com/visca-ai/visca/actions)

[Quickstart](#quickstart) | [Docs](https://docs.visca.ai) | [CLI Reference](#cli-reference) | [Installation](#installation)

VISCA enables organizations to set up RL training environments in their public or private cloud infrastructure. RL environments are defined with templates, connected through secure tunnels, and automatically shut down when not used to save on costs. VISCA gives engineering teams the flexibility to use the cloud for RL workloads most beneficial to them.

- Define RL training environments with templates
- Automatically shutdown idle resources to save on costs
- Onboard developers in seconds instead of days
- Version control environments including forking and sharing
- Distributed training across GPU clusters

## Quickstart

Install VISCA and experiment with provisioning RL training environments:

```shell
# Install VISCA
curl -L https://visca.ai/install.sh | sh

# Start the VISCA server
visca server

# Navigate to http://localhost:3000 to create your initial user,
# create a template and provision a workspace
```

## Installation

```shell
# Install script for Linux and macOS
curl -L https://visca.ai/install.sh | sh

# Start production deployment
visca server --postgres-url <url> --access-url <url>
```

Use `visca --help` to get a list of flags and environment variables.

## CLI Reference

### Authentication

```shell
# Login to VISCA deployment
visca login [url]

# Logout from current session
visca logout

# Show current user info
visca whoami
```

### Workspaces

```shell
# Create a new workspace
visca create [flags] [name]
visca create --template my-template my-workspace

# List workspaces
visca list
visca list --all

# Show workspace details
visca show <workspace>

# Start/stop workspaces
visca start <workspace>
visca stop <workspace>
visca restart <workspace>

# Delete workspace
visca delete <workspace>

# Update workspace
visca update <workspace>
```

### Templates

```shell
# List available templates
visca templates list

# Create template from current directory
visca templates push [template]

# Initialize with example template
visca templates init

# Pull template to local directory
visca templates pull <name> [destination]

# Edit template metadata
visca templates edit <template>

# Delete templates
visca templates delete [name...]
```

### SSH and Development

```shell
# SSH into workspace
visca ssh <workspace>

# Configure SSH
visca config-ssh

# Open workspace in VS Code
visca open vscode <workspace>

# Port forwarding
visca port-forward <workspace> --tcp 8080:8080
```

### Resource Management

```shell
# Show resource usage
visca stat
visca stat cpu
visca stat mem
visca stat disk

# Ping workspace
visca ping <workspace>

# Speed test
visca speedtest <workspace>
```

### Scheduling

```shell
# Show workspace schedules
visca schedule show <workspace>

# Set start schedule
visca schedule start <workspace> <time>

# Set stop schedule  
visca schedule stop <workspace> <duration>

# Override stop time
visca schedule override-stop <workspace> <duration>
```

### Organizations

```shell
# Show organization info
visca organizations show

# Create organization
visca organizations create <name>

# Manage members
visca organizations members list
visca organizations members add <username>
visca organizations members remove <username>
```

### Users

```shell
# List users
visca users list

# Create user
visca users create

# Show user details
visca users show <username>

# Activate/suspend users
visca users activate <username>
visca users suspend <username>
```

### Server Administration

```shell
# Start VISCA server
visca server [flags]

# Create admin user
visca server create-admin-user

# Database operations
visca server postgres-builtin-url
visca server postgres-builtin-serve

# Support bundle
visca support bundle <workspace>
```

### Tokens and Authentication

```shell
# Create API token
visca tokens create

# List tokens
visca tokens list

# Remove token
visca tokens remove <name>

# External authentication
visca external-auth access-token <provider>
```

### Utilities

```shell
# Network diagnostics
visca netcheck

# Version information
visca version

# Shell completion
visca completion

# Dotfiles setup
visca dotfiles <git_repo_url>

# Reset password
visca reset-password <username>
```

## Templates

Templates are written in Terraform and describe the infrastructure for workspaces. Templates define:

- Compute resources (CPU, memory, GPUs)
- Container images and software dependencies  
- Environment variables and configuration
- Networking and storage requirements

```hcl
# Example template
resource "visca_workspace" "dev" {
  name         = "rl-training"
  image        = "pytorch/pytorch:latest"
  cpu          = 4
  memory       = "16GB"
  gpu_count    = 1
  gpu_memory   = "12GB"
}
```

## Workspaces

Workspaces contain the IDEs, dependencies, and configuration information needed for RL development. Each workspace is:

- Isolated from other workspaces
- Automatically managed (start/stop/scale)
- Accessible via SSH, web IDEs, or desktop applications
- Version controlled with templates

## Self-Hosting

Deploy VISCA on your own infrastructure:

```shell
# Docker deployment
docker run -it --rm \
  -v ~/.config/visca:/home/visca/.config/visca \
  -p 3000:3000 \
  ghcr.io/visca-ai/visca:latest

# Kubernetes deployment
kubectl apply -f https://raw.githubusercontent.com/visca-ai/visca/main/install/kubernetes/

# Manual installation
visca server \
  --postgres-url postgresql://user:pass@localhost/visca \
  --access-url https://visca.company.com \
  --wildcard-access-url "*.visca.company.com"
```

## Documentation

- **[Installation](https://docs.visca.ai/install)**: Complete installation guide
- **[Templates](https://docs.visca.ai/templates)**: Creating and managing templates
- **[Workspaces](https://docs.visca.ai/workspaces)**: Workspace management
- **[CLI](https://docs.visca.ai/cli)**: Command line reference
- **[Administration](https://docs.visca.ai/admin)**: Server administration
- **[API](https://docs.visca.ai/api)**: REST API documentation

## Community

- **[Discord](https://discord.gg/v36hfGyJeN)**: Community chat and support
- **[GitHub Issues](https://github.com/visca-ai/visca/issues)**: Bug reports and feature requests
- **[GitHub Discussions](https://github.com/visca-ai/visca/discussions)**: Questions and ideas

## Contributing

We welcome contributions. See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

```shell
# Development setup
git clone https://github.com/visca-ai/visca.git
cd visca
make build
make install
```

## License

Licensed under the [MIT License](LICENSE).

---

**[Website](https://visca.ai) • [Platform](https://visca.dev) • [Documentation](https://docs.visca.ai) • [Community](https://discord.gg/v36hfGyJeN)**
