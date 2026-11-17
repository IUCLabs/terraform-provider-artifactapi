terraform {
  required_providers {
    artifactapi = {
      source = "IUCLabs/artifactapi"
    }
  }
}

provider "artifactapi" {
  endpoint = "https://artifactapi.example.com"
  # token defaults to the ARTIFACTAPI_TOKEN environment variable
}
