# Docker remote proxies a container registry.
# The provider auto-classifies: tag manifests are mutable, blobs are immutable.
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
