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

# Alpine remote proxies an Alpine Linux APK repository.
# The provider knows APKINDEX.tar.gz is mutable; .apk packages are immutable.
resource "artifactapi_remote_alpine" "alpine" {
  name        = "alpine"
  base_url    = "https://dl-cdn.alpinelinux.org"
  description = "Alpine Linux APK package repository"

  immutable_ttl = 0
  mutable_ttl   = 7200
}
