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

# A generic remote proxies arbitrary HTTP endpoints.
# Patterns act as an allowlist; everything matching is immutable by default.
resource "artifactapi_remote_generic" "github" {
  name        = "github"
  base_url    = "https://github.com"
  description = "GitHub releases"

  immutable_ttl = 0
  mutable_ttl   = 7200

  patterns = [
    "ducaale/xh/.*/xh-.*-x86_64-unknown-linux-musl.tar.gz$",
    "mikefarah/yq/.*/yq_linux_amd64$",
    "neovim/neovim-releases/.*/nvim-linux-x86_64.tar.gz$",
  ]

  # Override: branch archives are mutable
  mutable_patterns = [
    ".*/archive/refs/heads/.*\\.tar\\.gz$",
  ]
}
