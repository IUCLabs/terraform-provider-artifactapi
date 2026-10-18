# `artifactapi_local_ansible`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

A Galaxy v3 server for collections (`ansible-galaxy collection publish` and `install`).

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

See `examples/resources/artifactapi_local_ansible/main.tf`.
