# Helm remotes to be merged into a virtual repository
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

# Virtual repository merges multiple remotes of the same package type
# into a single endpoint. Earlier members have higher priority for
# duplicate entries.
resource "artifactapi_virtual" "helm" {
  name         = "helm"
  package_type = "helm"
  description  = "All Helm repos merged"

  members = [
    artifactapi_remote_helm.jetstack.name,
    artifactapi_remote_helm.hashicorp_helm.name,
  ]
}
