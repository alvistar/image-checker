# Kubernetes ImagePolicy Monitor

A monitoring tool that tracks container image updates across your Kubernetes cluster using FluxCD ImagePolicies and exposes Prometheus metrics.

## Features

- Monitors FluxCD ImagePolicies across all namespaces
- Checks if pods are running the latest available container images
- Exposes Prometheus metrics for tracking available updates
- Supports both in-cluster and local (development) execution
- Configurable polling interval
- Structured logging with timestamps

## Metrics

The tool exposes the following Prometheus metrics:

- `update_available`: A gauge metric indicating if a newer version is available (1) or not (0)
  - Labels:
    - `namespace`: Pod namespace
    - `pod`: Pod name
    - `container`: Container name

## Prerequisites

- Kubernetes cluster with FluxCD Image Reflector Controller installed
- Go 1.21 or later
- kubectl and kubeconfig (for local development)

## Installation

```bash
# Clone the repository
git clone [repository-url]
cd [repository-name]

# Install dependencies
go mod download
```

## Usage

### Running Locally

```bash
go run main.go [flags]
```

Available flags:
- `--interval`: Polling interval (default: 5m)
- `--listen-address`: Metrics server address (default: :2112)
- `--kubeconfig`: Path to kubeconfig file (optional)

### Running in Kubernetes

1. Create a ServiceAccount and necessary RBAC permissions
2. Deploy as a pod with appropriate configuration
3. Configure Prometheus to scrape the metrics endpoint

Example deployment configuration coming soon.

## Development

The tool is built using:
- client-go for Kubernetes API interaction
- controller-runtime for Kubernetes client operations
- Prometheus client for metrics exposure
- FluxCD Image Reflector Controller API for ImagePolicy types

## License

[Add your license information here]
