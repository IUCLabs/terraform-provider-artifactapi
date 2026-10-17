# terraform-provider-artifactapi

Terraform provider for managing [ArtifactAPI](https://git.unkin.net/unkin/artifactapi) remotes and virtual repositories.

## Requirements

- Go >= 1.23
- Terraform >= 1.0

## Building

```sh
make build
```

## Installation

Install the provider to your local Terraform plugin directory:

```sh
make install
```

This places the binary at `~/.terraform.d/plugins/git.unkin.net/unkin/artifactapi/<version>/<os_arch>/`.

## Provider Configuration

```hcl
terraform {
  required_providers {
    artifactapi = {
      source  = "git.unkin.net/unkin/artifactapi"
      version = "0.0.1"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.example.com"
  token    = var.artifactapi_token # optional
}
```

| Attribute  | Required | Description                                                                  |
|------------|----------|------------------------------------------------------------------------------|
| `endpoint` | Yes      | ArtifactAPI server endpoint URL                                              |
| `token`    | No       | Bearer token sent as `Authorization` header; falls back to `ARTIFACTAPI_TOKEN` |

## Resources and Data Sources

Each resource is documented in [`docs/resources/`](docs/resources/) and each data
source in [`docs/data-sources/`](docs/data-sources/), with a matching example under
[`examples/`](examples/).

## Import

Resources can be imported by name:

```sh
terraform import artifactapi_remote_docker.dockerhub dockerhub
terraform import artifactapi_virtual.helm helm
```

## Development

```sh
make build    # Build the provider binary
make install  # Install to local plugin directory
make test     # Run tests
make lint     # Run go vet
make fmt      # Format code
make clean    # Remove binary
```
