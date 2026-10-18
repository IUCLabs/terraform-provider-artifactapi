# `artifactapi_local_goproxy`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

A Go module proxy for private modules (GOPROXY protocol: `go get`, `go mod download`).

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

See `examples/resources/artifactapi_local_goproxy/main.tf`.
