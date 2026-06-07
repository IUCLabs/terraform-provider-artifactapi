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

# A generic remote is needed for the releases CDN that the Terraform registry
# redirects provider downloads to.
resource "artifactapi_remote_generic" "hashicorp_releases" {
  name        = "hashicorp-releases"
  base_url    = "https://releases.hashicorp.com"
  description = "HashiCorp product releases"

  immutable_ttl = 0
  mutable_ttl   = 7200

  patterns = [
    "terraform/.*terraform_.*_linux_amd64\\.zip$",
    "vault/.*vault_.*_linux_amd64\\.zip$",
  ]
}

# Terraform registry remote proxies the Terraform provider registry.
# The releases_remote attribute points to a generic remote that serves
# the actual provider/module archives after URL rewriting.
resource "artifactapi_remote_terraform" "terraform_registry" {
  name            = "terraform-registry"
  base_url        = "https://registry.terraform.io"
  description     = "Terraform provider registry"
  releases_remote = artifactapi_remote_generic.hashicorp_releases.name

  immutable_ttl = 0
  mutable_ttl   = 300
}
