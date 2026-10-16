terraform {
  required_providers {
    artifactapi = {
      source  = "git.unkin.net/unkin/artifactapi"
      version = "0.0.1"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.k8s.syd1.au.unkin.net"
  # token is read from ARTIFACTAPI_TOKEN when unset
}

# ---------------------------------------------------------------------------
# Generic remotes
# ---------------------------------------------------------------------------

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

  mutable_patterns = [
    ".*/archive/refs/heads/.*\\.tar\\.gz$",
  ]
}

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

# ---------------------------------------------------------------------------
# Docker
# ---------------------------------------------------------------------------

resource "artifactapi_remote_docker" "dockerhub" {
  name        = "dockerhub"
  base_url    = "https://registry-1.docker.io"
  description = "Docker Hub registry"

  immutable_ttl    = 0
  mutable_ttl      = 300
  ban_tags_enabled = true
  ban_tags         = ["latest"]

  patterns = [
    "^library/almalinux",
    "^library/postgres",
    "^library/redis",
    "^bitnami/",
  ]
}

# ---------------------------------------------------------------------------
# Helm
# ---------------------------------------------------------------------------

resource "artifactapi_remote_helm" "jetstack" {
  name        = "jetstack"
  base_url    = "https://charts.jetstack.io"
  description = "Jetstack Helm charts (cert-manager)"

  immutable_ttl = 0
  mutable_ttl   = 3600
}

resource "artifactapi_remote_helm" "hashicorp_helm" {
  name        = "hashicorp-helm"
  base_url    = "https://helm.releases.hashicorp.com"
  description = "HashiCorp Helm charts"

  immutable_ttl = 0
  mutable_ttl   = 3600
}

# ---------------------------------------------------------------------------
# Language package managers
# ---------------------------------------------------------------------------

resource "artifactapi_remote_pypi" "pypi" {
  name        = "pypi"
  base_url    = "https://pypi.org"
  description = "Python Package Index"

  immutable_ttl = 0
  mutable_ttl   = 900
}

resource "artifactapi_remote_npm" "npmjs" {
  name        = "npmjs"
  base_url    = "https://registry.npmjs.org"
  description = "npm public registry"

  immutable_ttl = 0
  mutable_ttl   = 900
}

# ---------------------------------------------------------------------------
# OS package managers
# ---------------------------------------------------------------------------

resource "artifactapi_remote_rpm" "almalinux" {
  name        = "almalinux"
  base_url    = "https://gsl-syd.mm.fcix.net/almalinux"
  description = "AlmaLinux RPM package repository"

  immutable_ttl = 0
  mutable_ttl   = 7200
}

resource "artifactapi_remote_alpine" "alpine" {
  name        = "alpine"
  base_url    = "https://dl-cdn.alpinelinux.org"
  description = "Alpine Linux APK package repository"

  immutable_ttl = 0
  mutable_ttl   = 7200
}

# ---------------------------------------------------------------------------
# Infrastructure tooling
# ---------------------------------------------------------------------------

resource "artifactapi_remote_terraform" "terraform_registry" {
  name            = "terraform-registry"
  base_url        = "https://registry.terraform.io"
  description     = "Terraform provider registry"
  releases_remote = artifactapi_remote_generic.hashicorp_releases.name

  immutable_ttl = 0
  mutable_ttl   = 300
}

resource "artifactapi_remote_goproxy" "goproxy" {
  name        = "goproxy"
  base_url    = "https://proxy.golang.org"
  description = "Go module proxy"

  immutable_ttl = 0
  mutable_ttl   = 300
}

resource "artifactapi_remote_puppet" "puppet_forge" {
  name        = "puppet-forge"
  base_url    = "https://forgeapi.puppet.com"
  description = "Puppet Forge module repository"

  immutable_ttl = 0
  mutable_ttl   = 3600
}

# ---------------------------------------------------------------------------
# Virtual repositories
# ---------------------------------------------------------------------------

resource "artifactapi_virtual" "helm" {
  name         = "helm"
  package_type = "helm"
  description  = "All Helm repos merged"

  members = [
    artifactapi_remote_helm.jetstack.name,
    artifactapi_remote_helm.hashicorp_helm.name,
  ]
}

# ---------------------------------------------------------------------------
# Data sources
# ---------------------------------------------------------------------------

data "artifactapi_remote" "dockerhub" {
  name = artifactapi_remote_docker.dockerhub.name
}

data "artifactapi_virtual" "helm" {
  name = artifactapi_virtual.helm.name
}

output "dockerhub_base_url" {
  value = data.artifactapi_remote.dockerhub.base_url
}

output "helm_members" {
  value = data.artifactapi_virtual.helm.members
}
