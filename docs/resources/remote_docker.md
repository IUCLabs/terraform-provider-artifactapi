# `artifactapi_remote_docker`

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

## Docker-specific Attributes

| Attribute          | Default | Description                |
|--------------------|---------|----------------------------|
| `ban_tags_enabled` | `false` | Enable tag banning         |
| `ban_tags`         |         | List of tags to ban        |

## Example

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
