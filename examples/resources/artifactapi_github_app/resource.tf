# GitHub App credential; remotes reference it by name to authenticate with
# installation tokens instead of a static token in `password`.
resource "artifactapi_github_app" "acme" {
  name            = "acme"
  app_id          = 123456
  installation_id = 7654321
  private_key     = file("acme-app.private-key.pem")
  description     = "acme GitHub App"

  # Optional: GitHub Enterprise Server API root (default https://api.github.com).
  # api_base_url = "https://ghe.example.com/api/v3"
}

resource "artifactapi_remote_github_rpm" "acme_tools" {
  name            = "acme-tools"
  base_url        = "https://api.github.com/repos/acme/tools"
  releases_remote = "github"
  github_app      = artifactapi_github_app.acme.name
}
