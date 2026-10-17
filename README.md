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
- `artifactapi_remote_deb`
- `artifactapi_remote_alpine`
- `artifactapi_remote_puppet`
- `artifactapi_remote_terraform`
- `artifactapi_remote_goproxy`
- `artifactapi_remote_cargo`
- `artifactapi_remote_gems`
- `artifactapi_remote_github_rpm`
- `artifactapi_remote_github_deb`
- `artifactapi_remote_github_alpine`

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

#### rpm / deb / alpine-specific Attributes

| Attribute         | Default       | Description                                                                                     |
|-------------------|---------------|-------------------------------------------------------------------------------------------------|
| `mirrorlist`      |               | Extra upstream mirror base URLs. Requests are load-balanced with failover across `base_url` + `mirrorlist`. |
| `mirror_strategy` | `round_robin` | Mirror load-balancing strategy across `base_url` + `mirrorlist`. One of `round_robin` or `least_conn`. |

Only valid on the `artifactapi_remote_rpm`, `artifactapi_remote_deb`, and
`artifactapi_remote_alpine` resources. Setting either on any other remote type is
rejected at plan time.

```hcl
resource "artifactapi_remote_rpm" "epel" {
  name     = "epel-9"
  base_url = "https://download.example.com/pub/epel/9/Everything/x86_64"
  mirrorlist = [
    "https://mirror-a.example.net/epel/9/Everything/x86_64",
    "https://mirror-b.example.org/epel/9/Everything/x86_64",
  ]
  mirror_strategy = "least_conn"
}
```

#### Docker-specific Attributes

| Attribute          | Default | Description                |
|--------------------|---------|----------------------------|
| `ban_tags_enabled` | `false` | Enable tag banning         |
| `ban_tags`         |         | List of tags to ban        |

#### Terraform / github_rpm / github_deb / github_alpine-specific Attributes

| Attribute         | Default | Description                                              |
|-------------------|---------|----------------------------------------------------------|
| `releases_remote` | `""`   | Name of a backend remote that serves the download bytes. Terraform uses it for download URL rewriting; `github_rpm`/`github_deb`/`github_alpine` 302-redirect `.rpm`/`.deb`/`.apk` downloads to it. |

#### github_rpm-specific notes

`artifactapi_remote_github_rpm` exposes a GitHub repo's releases as a `dnf`/`yum`
repository without precaching packages. Set `base_url` to the releases API root
(`https://api.github.com/repos/{owner}/{repo}`) and `releases_remote` to a
generic `github.com` remote that streams the `.rpm` bytes. `patterns` filters
which release assets become packages (regex on asset filename). `password` may
hold a token for private repos or higher API rate limits. See
`examples/resources/artifactapi_remote_github_rpm/main.tf`.

#### github_deb-specific notes

`artifactapi_remote_github_deb` is the Debian/apt analog of `github_rpm`: it
exposes a GitHub repo's release `.deb` assets as an `apt` repository, synthesizing
the apt index (`Packages`/`Release`) from release metadata without precaching
packages. Set `base_url` to the releases API root
(`https://api.github.com/repos/{owner}/{repo}`) and `releases_remote` to a
generic `github.com` remote that streams the `.deb` bytes; `.deb` downloads
302-redirect there. `patterns` filters which release assets become packages
(regex on asset filename). `password` may hold a token for private repos or
higher API rate limits. See
`examples/resources/artifactapi_remote_github_deb/main.tf`.

#### github_alpine-specific notes

`artifactapi_remote_github_alpine` is the Alpine/apk analog of `github_rpm`: it
exposes a GitHub repo's release `.apk` assets as an `apk` repository, synthesizing
the apk index (`APKINDEX.tar.gz`) from release metadata without precaching
packages. Set `base_url` to the releases API root
(`https://api.github.com/repos/{owner}/{repo}`) and `releases_remote` to a
generic `github.com` remote that streams the `.apk` bytes; `.apk` downloads
302-redirect there. `patterns` filters which release assets become packages
(regex on asset filename). `password` may hold a token for private repos or
higher API rate limits. See
`examples/resources/artifactapi_remote_github_alpine/main.tf`.

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
- `artifactapi_local_npm` — an npm registry (`npm publish` and `npm install`)
- `artifactapi_local_rpm`
- `artifactapi_local_deb`
- `artifactapi_local_alpine`
- `artifactapi_local_puppet` — a Puppet Forge (v3 API for `puppet module install`, r10k and g10k)
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
