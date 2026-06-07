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

# PyPI remote proxies a Python package index.
# The provider knows /simple/ index pages are mutable; sdists and wheels are immutable.
resource "artifactapi_remote_pypi" "pypi" {
  name        = "pypi"
  base_url    = "https://pypi.org"
  description = "Python Package Index"

  immutable_ttl = 0
  mutable_ttl   = 900
}
