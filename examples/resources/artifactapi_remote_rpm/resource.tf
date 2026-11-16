# RPM remote proxies an RPM package repository.
# The provider knows repodata/* is mutable; RPM packages are immutable.
resource "artifactapi_remote_rpm" "almalinux" {
  name        = "almalinux"
  base_url    = "https://repo.almalinux.org/almalinux"
  description = "AlmaLinux RPM package repository"

  # Extra mirrors; requests are load-balanced with failover across
  # base_url + mirrorlist (rpm/deb/alpine remotes only).
  mirrorlist = [
    "https://mirror-a.example.net/almalinux",
    "https://mirror-b.example.org/almalinux",
  ]

  # Load-balancing strategy across base_url + mirrorlist:
  # "round_robin" (default) or "least_conn".
  mirror_strategy = "least_conn"

  immutable_ttl = 0
  mutable_ttl   = 7200
}
