# Backend generic remote that serves the actual .apk bytes from github.com.
# Its patterns must allowlist the target repo's release-download assets.
resource "artifactapi_remote_generic" "github" {
  name     = "github"
  base_url = "https://github.com"

  patterns = [
    "acme/tools/releases/download/.*\\.apk$",
  ]
}

# Metadata-only remote: scans acme/tools releases for .apk assets and serves a
# synthesized apk repo (APKINDEX). Packages are never precached; downloads
# 302-redirect to the "github" remote above.
resource "artifactapi_remote_github_alpine" "acme_tools" {
  name            = "acme-tools"
  base_url        = "https://api.github.com/repos/acme/tools"
  description     = "acme/tools GitHub releases as an apk repo"
  releases_remote = artifactapi_remote_generic.github.name

  mutable_ttl = 3600

  # Optional: restrict which release assets become packages.
  patterns = [
    ".*_x86_64\\.apk$",
    ".*_noarch\\.apk$",
  ]

  # Optional: token for private repos / higher GitHub API rate limits.
  # password = var.github_token
  # Or authenticate as a GitHub App (conflicts with username/password).
  # github_app = artifactapi_github_app.acme.name
}
