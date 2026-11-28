# GitHub App credential; remotes reference it by name to authenticate with
# installation tokens instead of a static token in `password`.
resource "artifactapi_github_app" "acme" {
  name            = "acme"
  app_id          = 123456
  installation_id = 7654321
  description     = "acme GitHub App"

  # Write-only (Terraform 1.11+): never stored in plan or state. The key can
  # also come from an ephemeral resource such as ephemeral "vault_kv_secret_v2":
  #   private_key_wo = ephemeral.vault_kv_secret_v2.acme.data["private_key"]
  private_key_wo = file("acme-app.private-key.pem")
  # Bump to resend the key; a key whose fingerprint differs from the server's
  # is resent automatically.
  private_key_wo_version = 1

  # Terraform < 1.11: use private_key instead (stored in state).
  # private_key = file("acme-app.private-key.pem")

  # Optional: GitHub Enterprise Server API root (default https://api.github.com).
  # api_base_url = "https://ghe.example.com/api/v3"
}

resource "artifactapi_remote_github_rpm" "acme_tools" {
  name            = "acme-tools"
  base_url        = "https://api.github.com/repos/acme/tools"
  releases_remote = "github"
  github_app      = artifactapi_github_app.acme.name
}
