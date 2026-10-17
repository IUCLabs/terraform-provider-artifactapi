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

# Gems remote proxies a RubyGems source.
# The provider knows the compact index is mutable; .gem downloads are immutable.
resource "artifactapi_remote_gems" "rubygems" {
  name        = "rubygems"
  base_url    = "https://rubygems.org"
  description = "rubygems.org"

  immutable_ttl = 0
  mutable_ttl   = 300
}
