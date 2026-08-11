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

# Deb remote proxies a Debian/apt package repository.
# The provider knows the index files (Release, Packages) are mutable; .deb packages are immutable.
resource "artifactapi_remote_deb" "debian" {
  name        = "debian"
  base_url    = "http://deb.debian.org/debian"
  description = "Debian apt package repository"

  immutable_ttl = 0
  mutable_ttl   = 7200
}
