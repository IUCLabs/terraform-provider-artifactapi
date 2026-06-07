terraform {
  required_providers {
    artifactapi = {
      source = "git.unkin.net/unkin/artifactapi"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.k8s.syd1.au.unkin.net"
}

resource "artifactapi_remote" "dockerhub" {
  name         = "dockerhub"
  package_type = "docker"
  base_url     = "https://registry-1.docker.io"
  description  = "Docker Hub registry"

  immutable_ttl = 0
  mutable_ttl   = 300

  immutable_patterns = [
    "^library/almalinux",
    "^library/postgres",
    "^library/redis",
  ]
}

resource "artifactapi_remote" "hashicorp_releases" {
  name         = "hashicorp-releases"
  package_type = "generic"
  base_url     = "https://releases.hashicorp.com"
  description  = "HashiCorp product releases"

  immutable_ttl = 0
  mutable_ttl   = 7200

  immutable_patterns = [
    "terraform/.*terraform_.*_linux_amd64\\.zip$",
    "vault/.*vault_.*_linux_amd64\\.zip$",
  ]
}

resource "artifactapi_remote" "terraform_registry" {
  name            = "terraform-registry"
  package_type    = "terraform"
  base_url        = "https://registry.terraform.io"
  description     = "Terraform provider registry"
  releases_remote = artifactapi_remote.hashicorp_releases.name

  immutable_ttl = 0
  mutable_ttl   = 300

  immutable_patterns = [
    "[^/]+/[^/]+/[^/]+/download/[^/]+/[^/]+$",
  ]
}

resource "artifactapi_remote" "goproxy" {
  name         = "goproxy"
  package_type = "goproxy"
  base_url     = "https://proxy.golang.org"
  description  = "Go module proxy"

  immutable_ttl = 0
  mutable_ttl   = 300
}

resource "artifactapi_remote" "jetstack" {
  name         = "jetstack"
  package_type = "helm"
  base_url     = "https://charts.jetstack.io"
  description  = "Jetstack Helm charts (cert-manager)"

  immutable_ttl  = 0
  mutable_ttl    = 3600
  check_mutable  = true

  immutable_patterns = ["\\.tgz$"]
}

resource "artifactapi_remote" "hashicorp_helm" {
  name         = "hashicorp-helm"
  package_type = "helm"
  base_url     = "https://helm.releases.hashicorp.com"
  description  = "HashiCorp Helm charts"

  immutable_ttl  = 0
  mutable_ttl    = 3600
  check_mutable  = true

  immutable_patterns = ["\\.tgz$"]
}

resource "artifactapi_virtual" "helm" {
  name         = "helm"
  package_type = "helm"
  description  = "All helm repos merged"

  members = [
    artifactapi_remote.jetstack.name,
    artifactapi_remote.hashicorp_helm.name,
  ]
}

data "artifactapi_remote" "dockerhub" {
  name = artifactapi_remote.dockerhub.name
}

output "dockerhub_base_url" {
  value = data.artifactapi_remote.dockerhub.base_url
}
