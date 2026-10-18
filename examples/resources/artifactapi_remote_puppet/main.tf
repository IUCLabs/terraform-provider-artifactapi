terraform {
  required_providers {
    artifactapi = {
      source  = "artifactapi.example.com/terraform-unkin/artifactapi"
      version = "0.0.1"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.example.com"
}

# Puppet remote proxies a Puppet Forge repository.
# The provider knows module metadata is mutable; release tarballs are immutable.
resource "artifactapi_remote_puppet" "puppet_forge" {
  name        = "puppet-forge"
  base_url    = "https://forgeapi.puppet.com"
  description = "Puppet Forge module repository"

  immutable_ttl = 0
  mutable_ttl   = 3600
}
