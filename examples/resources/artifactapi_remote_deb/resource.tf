# Deb remote proxies a Debian/apt package repository.
# The provider knows the index files (Release, Packages) are mutable; .deb packages are immutable.
resource "artifactapi_remote_deb" "debian" {
  name        = "debian"
  base_url    = "http://deb.debian.org/debian"
  description = "Debian apt package repository"

  immutable_ttl = 0
  mutable_ttl   = 7200
}
