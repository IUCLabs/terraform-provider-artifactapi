# `artifactapi_local_gems`

Local resources manage repositories that ArtifactAPI hosts directly (rather than
proxying an upstream) — each is a real registry for its package type.

A RubyGems source (`gem push`, `gem yank`, `gem install` and `bundle install`).

Each takes just `name` (required, forces replacement) and an optional
`description`.

## Example

See `examples/resources/artifactapi_local_gems/main.tf`.
