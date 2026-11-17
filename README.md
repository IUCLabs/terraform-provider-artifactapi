# Terraform Provider ArtifactAPI

The ArtifactAPI provider manages repositories on an ArtifactAPI server: local
repositories that host packages directly, remotes that cache an upstream
(Docker, Helm, PyPI, npm, RPM, Deb, Alpine, Go, Terraform, GitHub releases and
more), and virtual repositories that merge several remotes behind one endpoint.
It works with Terraform and OpenTofu.

## Quick start

```hcl
terraform {
  required_providers {
    artifactapi = {
      source = "IUCLabs/artifactapi"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.example.com"
  # token defaults to the ARTIFACTAPI_TOKEN environment variable
}

resource "artifactapi_remote_docker" "dockerhub" {
  name        = "dockerhub"
  base_url    = "https://registry-1.docker.io"
  mutable_ttl = 300
}

resource "artifactapi_local_rpm" "internal" {
  name = "rpm-internal"
}
```

All resources can be imported by name, e.g.
`terraform import artifactapi_remote_docker.dockerhub dockerhub`.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0 or [OpenTofu](https://opentofu.org) >= 1.6
- [Go](https://go.dev/doc/install) >= 1.25 (to build the provider)

## Building

```sh
make build
make test
make docs    # regenerate docs/ from the schema and examples/
```

## Documentation

See [`docs/`](docs/index.md) for the provider, resource and data source
reference, and [`examples/`](examples/) for usage.

## License

[MIT](LICENSE)
