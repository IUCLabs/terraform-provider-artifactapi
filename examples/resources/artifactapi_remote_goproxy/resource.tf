# Go module proxy remote proxies a Go module proxy.
# The provider knows @v/list is mutable; .zip, .mod, and .info files are immutable.
resource "artifactapi_remote_goproxy" "goproxy" {
  name        = "goproxy"
  base_url    = "https://proxy.golang.org"
  description = "Go module proxy"

  immutable_ttl = 0
  mutable_ttl   = 300
}
