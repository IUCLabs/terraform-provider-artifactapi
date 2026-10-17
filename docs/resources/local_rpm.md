# `artifactapi_local_rpm`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

See `examples/resources/artifactapi_local_rpm/main.tf`.
