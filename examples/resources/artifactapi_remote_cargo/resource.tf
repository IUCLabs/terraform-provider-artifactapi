# Cargo remote proxies a crates.io sparse registry.
# The provider knows index files are mutable; .crate downloads are immutable.
resource "artifactapi_remote_cargo" "crates_io" {
  name        = "crates-io"
  base_url    = "https://index.crates.io"
  description = "crates.io sparse registry"

  immutable_ttl = 0
  mutable_ttl   = 300
}
