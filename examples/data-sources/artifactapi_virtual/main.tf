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

# Read the configuration of an existing virtual repository by name.
# Returns the package type, description, and ordered member list.
data "artifactapi_virtual" "helm" {
  name = "helm"
}

output "helm_members" {
  value = data.artifactapi_virtual.helm.members
}

output "helm_package_type" {
  value = data.artifactapi_virtual.helm.package_type
}
