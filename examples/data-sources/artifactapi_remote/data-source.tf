# Read the configuration of an existing remote by name.
# All fields (base_url, package_type, TTLs, patterns, etc.) are available
# as computed attributes.
data "artifactapi_remote" "dockerhub" {
  name = "dockerhub"
}

output "dockerhub_base_url" {
  value = data.artifactapi_remote.dockerhub.base_url
}

output "dockerhub_package_type" {
  value = data.artifactapi_remote.dockerhub.package_type
}
