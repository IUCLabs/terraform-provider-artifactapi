# Backend generic remote that serves the actual .deb bytes from github.com.
# Its patterns must allowlist the target repo's release-download assets.
resource "artifactapi_remote_generic" "github" {
  name     = "github"
  base_url = "https://github.com"

  patterns = [
    "acme/tools/releases/download/.*\\.deb$",
  ]
}

# Metadata-only remote: scans acme/tools releases for .deb assets and serves a
# synthesized apt repo. Packages are never precached; downloads 302-redirect to
# the "github" remote above.
resource "artifactapi_remote_github_deb" "acme_tools" {
  name            = "acme-tools"
  base_url        = "https://api.github.com/repos/acme/tools"
  description     = "acme/tools GitHub releases as an apt repo"
  releases_remote = artifactapi_remote_generic.github.name

  mutable_ttl = 3600

  # Optional: restrict which release assets become packages.
  patterns = [
    ".*_amd64\\.deb$",
    ".*_all\\.deb$",
  ]

  # Optional: token for private repos / higher GitHub API rate limits.
  # password = var.github_token
}
