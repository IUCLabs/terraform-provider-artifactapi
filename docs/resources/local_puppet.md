# `artifactapi_local_puppet`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

A Puppet Forge (v3 API) serving uploaded modules to `puppet module install` and r10k.

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

See `examples/resources/artifactapi_local_puppet/main.tf`.
