terraform {
  required_providers {
    artifactapi = {
      source  = "git.unkin.net/unkin/artifactapi"
      version = "0.0.1"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.example.com"
}

# Backend generic remote that serves the actual .deb bytes from github.com.
# Its patterns must allowlist the target repo's release-download assets.
resource "artifactapi_remote_generic" "github" {
  name     = "github"
  base_url = "https://github.com"

  patterns = [
    "goodtune/ghp/releases/download/.*\\.deb$",
  ]
}

# Metadata-only remote: scans goodtune/ghp releases for .deb assets and serves a
# synthesized apt repo. Packages are never precached; downloads 302-redirect to
# the "github" remote above.
resource "artifactapi_remote_github_deb" "goodtune_ghp" {
  name            = "goodtune-ghp"
  base_url        = "https://api.github.com/repos/goodtune/ghp"
  description     = "goodtune/ghp GitHub releases as an apt repo"
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
