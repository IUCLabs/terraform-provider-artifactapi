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

# Ansible remote proxies the Galaxy v3 collections API.
# The provider knows JSON metadata is mutable; collection tarballs are immutable.
resource "artifactapi_remote_ansible" "galaxy" {
  name        = "galaxy"
  base_url    = "https://galaxy.ansible.com"
  description = "Ansible Galaxy"

  immutable_ttl = 0
  mutable_ttl   = 300
}
