# `artifactapi_local_helm`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

A chart repository (index.yaml generated from uploaded charts).

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

See `examples/resources/artifactapi_local_helm/main.tf`.
