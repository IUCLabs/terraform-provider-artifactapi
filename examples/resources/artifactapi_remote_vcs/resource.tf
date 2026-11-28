# VCS remote caches GitHub/Gitea source archives and release assets.
# Tag, commit and release paths are immutable; branch archives redirect to
# the archive of the branch's current commit, re-resolved after mutable_ttl.
resource "artifactapi_remote_vcs" "github" {
  name        = "github-vcs"
  base_url    = "https://github.com"
  description = "GitHub source archives and release assets"

  immutable_ttl = 0
  mutable_ttl   = 600

  patterns = ["^cli/cli/", "^psf/requests/"]

  # Optional: authenticate as a GitHub App for private repos / higher rate limits.
  # github_app = artifactapi_github_app.acme.name
}
