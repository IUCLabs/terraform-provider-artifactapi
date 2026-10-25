# `artifactapi_remote_rpm`

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
| `serve_stale`        | No       | `false` | Let rpm/deb/alpine virtuals serve this member's last good index while it refreshes |
| `upstream_dial_timeout` | No    | `0`     | Upstream TCP connect timeout in seconds (0 = server default)      |
| `upstream_tls_timeout` | No     | `0`     | Upstream TLS handshake timeout in seconds (0 = server default)    |
| `upstream_response_header_timeout` | No | `0` | Upstream response-header timeout in seconds (0 = server default) |

## rpm / deb / alpine-specific Attributes

| Attribute         | Default       | Description                                                                                     |
|-------------------|---------------|-------------------------------------------------------------------------------------------------|
| `mirrorlist`      |               | Extra upstream mirror base URLs. Requests are load-balanced with failover across `base_url` + `mirrorlist`. |
| `mirror_strategy` | `round_robin` | Mirror load-balancing strategy across `base_url` + `mirrorlist`. One of `round_robin` or `least_conn`. |

Only valid on the `artifactapi_remote_rpm`, `artifactapi_remote_deb`, and
`artifactapi_remote_alpine` resources. Setting either on any other remote type is
rejected at plan time.

```hcl
resource "artifactapi_remote_rpm" "epel" {
  name     = "epel-9"
  base_url = "https://download.example.com/pub/epel/9/Everything/x86_64"
  mirrorlist = [
    "https://mirror-a.example.net/epel/9/Everything/x86_64",
    "https://mirror-b.example.org/epel/9/Everything/x86_64",
  ]
  mirror_strategy = "least_conn"
}
```

## Example

See `examples/resources/artifactapi_remote_rpm/main.tf`.
