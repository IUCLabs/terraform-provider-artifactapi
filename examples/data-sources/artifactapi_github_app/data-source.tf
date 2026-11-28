# Reference a GitHub App created in the ArtifactAPI UI; its private key stays on the server.
data "artifactapi_github_app" "acme" {
  name = "acme"
}

resource "artifactapi_remote_github_rpm" "acme_tools" {
  name            = "acme-tools"
  base_url        = "https://api.github.com/repos/acme/tools"
  releases_remote = "github"
  github_app      = data.artifactapi_github_app.acme.name
}
