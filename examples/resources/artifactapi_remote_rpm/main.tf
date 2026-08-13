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

# RPM remote proxies an RPM package repository.
# The provider knows repodata/* is mutable; RPM packages are immutable.
resource "artifactapi_remote_rpm" "almalinux" {
  name        = "almalinux"
  base_url    = "https://gsl-syd.mm.fcix.net/almalinux"
  description = "AlmaLinux RPM package repository"

  # Extra mirrors; requests are load-balanced with failover across
  # base_url + mirrorlist (rpm/deb/alpine remotes only).
  mirrorlist = [
    "https://mirror.aarnet.edu.au/pub/almalinux",
    "https://mirror.realcompute.io/almalinux",
  ]

  immutable_ttl = 0
  mutable_ttl   = 7200
}
