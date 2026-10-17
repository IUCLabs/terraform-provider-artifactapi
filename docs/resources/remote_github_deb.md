# `artifactapi_remote_github_deb`

Per-type remote resources manage upstream repository proxies. Each type applies its own mutability classification rules automatically (e.g., Docker classifies tag manifests as mutable and blobs as immutable; Helm classifies `index.yaml` as mutable).

## Common Attributes

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

## Terraform / github_rpm / github_deb / github_alpine-specific Attributes

| Attribute         | Default | Description                                              |
|-------------------|---------|----------------------------------------------------------|
| `releases_remote` | `""`   | Name of a backend remote that serves the download bytes. Terraform uses it for download URL rewriting; `github_rpm`/`github_deb`/`github_alpine` 302-redirect `.rpm`/`.deb`/`.apk` downloads to it. |

## Notes

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

## Example

See `examples/resources/artifactapi_remote_github_deb/main.tf`.
