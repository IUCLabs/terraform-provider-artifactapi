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

# NPM remote proxies an npm registry.
# The provider knows package metadata is mutable; tarballs are immutable.
resource "artifactapi_remote_npm" "npmjs" {
  name        = "npmjs"
  base_url    = "https://registry.npmjs.org"
  description = "npm public registry"

  immutable_ttl = 0
  mutable_ttl   = 900
}
