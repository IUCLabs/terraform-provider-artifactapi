# Backend generic remote that serves the actual .rpm bytes from github.com.
# Its patterns must allowlist the target repo's release-download assets.
resource "artifactapi_remote_generic" "github" {
  name     = "github"
  base_url = "https://github.com"

  patterns = [
    "acme/tools/releases/download/.*\\.rpm$",
  ]
}

# Metadata-only remote: scans acme/tools releases for .rpm assets and serves a
# synthesized yum repo. Packages are never precached; downloads 302-redirect to
# the "github" remote above.
resource "artifactapi_remote_github_rpm" "acme_tools" {
  name            = "acme-tools"
  base_url        = "https://api.github.com/repos/acme/tools"
  description     = "acme/tools GitHub releases as a yum repo"
  releases_remote = artifactapi_remote_generic.github.name

  mutable_ttl = 3600

  # Optional: re-scan releases every 15 minutes, cache the generated repodata,
  # and let virtuals serve the last good index while this remote refreshes.
  sync_interval  = 900
  cache_repodata = true
  serve_stale    = true

  # Optional: restrict which release assets become packages.
  patterns = [
    ".*\\.x86_64\\.rpm$",
    ".*\\.noarch\\.rpm$",
  ]

  # Optional: token for private repos / higher GitHub API rate limits.
  # password = var.github_token
}
