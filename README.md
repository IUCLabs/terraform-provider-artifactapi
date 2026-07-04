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
}
```

| Attribute  | Required | Description                          |
|------------|----------|--------------------------------------|
| `endpoint` | Yes      | ArtifactAPI server endpoint URL      |

## Resources

### Remote Resources

Per-type remote resources manage upstream repository proxies. Each type applies its own mutability classification rules automatically (e.g., Docker classifies tag manifests as mutable and blobs as immutable; Helm classifies `index.yaml` as mutable).

Available resource types:

- `artifactapi_remote_generic`
- `artifactapi_remote_docker`
- `artifactapi_remote_helm`
- `artifactapi_remote_pypi`
- `artifactapi_remote_npm`
- `artifactapi_remote_rpm`
- `artifactapi_remote_alpine`
- `artifactapi_remote_puppet`
- `artifactapi_remote_terraform`
- `artifactapi_remote_goproxy`

#### Common Attributes

| Attribute            | Required | Default | Description                                                       |
|----------------------|----------|---------|-------------------------------------------------------------------|
| `name`               | Yes      |         | Unique name (forces replacement on change)                        |
| `base_url`           | Yes      |         | Upstream repository URL                                           |
| `description`        | No       | `""`    | Human-readable description                                        |
| `username`           | No       | `""`    | Upstream auth username (sensitive)                                 |
| `password`           | No       | `""`    | Upstream auth password (sensitive)                                 |
| `immutable_ttl`      | No       | `0`     | TTL in seconds for immutable artifacts (0 = cache forever)        |
| `mutable_ttl`        | No       | `3600`  | TTL in seconds for mutable artifacts                              |
| `check_mutable`      | No       | `true`  | Enable conditional revalidation for mutable artifacts             |
| `patterns`           | No       |         | Allowlist of path patterns to proxy (empty = all)                 |
| `blocklist`          | No       |         | Paths to always deny (checked before patterns)                    |
| `mutable_patterns`   | No       |         | Override: treat matching paths as mutable                         |
| `immutable_patterns` | No       |         | Override: treat matching paths as immutable                       |
| `quarantine_enabled` | No       | `false` | Enable quarantine for new artifacts                               |
| `quarantine_days`    | No       | `3`     | Days to quarantine new artifacts                                  |
| `stale_on_error`     | No       | `true`  | Serve stale cache when upstream is unreachable                    |
| `upstream_dial_timeout` | No    | `0`     | Upstream TCP connect timeout in seconds (0 = server default)      |
| `upstream_tls_timeout` | No     | `0`     | Upstream TLS handshake timeout in seconds (0 = server default)    |
| `upstream_response_header_timeout` | No | `0` | Upstream response-header timeout in seconds (0 = server default) |

#### Docker-specific Attributes

| Attribute          | Default | Description                |
|--------------------|---------|----------------------------|
| `ban_tags_enabled` | `false` | Enable tag banning         |
| `ban_tags`         |         | List of tags to ban        |

#### Terraform-specific Attributes

| Attribute         | Default | Description                                              |
|-------------------|---------|----------------------------------------------------------|
| `releases_remote` | `""`   | Name of a generic remote for download URL rewriting      |

#### Example

```hcl
resource "artifactapi_remote_docker" "dockerhub" {
  name     = "dockerhub"
  base_url = "https://registry-1.docker.io"

  immutable_ttl    = 0
  mutable_ttl      = 300
  ban_tags_enabled = true
  ban_tags         = ["latest"]

  patterns = [
    "^library/postgres",
    "^library/redis",
  ]
}
```

### Local Resources

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

Available resource types:

- `artifactapi_local_docker` — a container registry (Docker Registry HTTP API V2, push and pull)
- `artifactapi_local_pypi`
- `artifactapi_local_rpm`
- `artifactapi_local_terraform`

Each takes just `name` (required, forces replacement) and an optional
`description`.

```hcl
resource "artifactapi_local_docker" "internal" {
  name        = "docker-internal"
  description = "Internal container image registry"
}
```

Images push and pull against `<endpoint>/<name>/<image>:<tag>`, e.g.
`docker push artifactapi.example.com/docker-internal/myapp:latest`.

### Virtual Resources

Virtual repositories merge multiple remotes of the same package type into a single endpoint.

```hcl
resource "artifactapi_virtual" "helm" {
  name         = "helm"
  package_type = "helm"
  description  = "All helm repos merged"

  members = [
    artifactapi_remote_helm.jetstack.name,
    artifactapi_remote_helm.hashicorp_helm.name,
  ]
}
```

| Attribute      | Required | Description                               |
|----------------|----------|-------------------------------------------|
| `name`         | Yes      | Unique name (forces replacement on change)|
| `package_type` | Yes      | Package type of member remotes            |
| `description`  | No       | Human-readable description                |
| `members`      | Yes      | List of remote names to include           |

## Data Sources

### `artifactapi_remote`

Read an existing remote's configuration.

```hcl
data "artifactapi_remote" "dockerhub" {
  name = "dockerhub"
}
```

### `artifactapi_virtual`

Read an existing virtual repository's configuration.

```hcl
data "artifactapi_virtual" "helm" {
  name = "helm"
}
```

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
