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

# Helm remote proxies a Helm chart repository.
# The provider knows index.yaml is mutable; chart tarballs are immutable.
resource "artifactapi_remote_helm" "jetstack" {
  name        = "jetstack"
  base_url    = "https://charts.jetstack.io"
  description = "Jetstack Helm charts (cert-manager)"

  immutable_ttl = 0
  mutable_ttl   = 3600
}
