# `artifactapi_local_docker`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

A container registry (Docker Registry HTTP API V2, push and pull).

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

```hcl
resource "artifactapi_local_docker" "internal" {
  name        = "docker-internal"
  description = "Internal container image registry"
}
```

Images push and pull against `<endpoint>/<name>/<image>:<tag>`, e.g.
`docker push artifactapi.example.com/docker-internal/myapp:latest`.
