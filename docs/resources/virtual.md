# `artifactapi_virtual`

Virtual repositories merge multiple remotes of the same package type into a single endpoint.

```hcl
resource "artifactapi_virtual" "helm" {
  name         = "helm"
  package_type = "helm"
  description  = "All helm repos merged"

  members = [
    artifactapi_remote_helm.jetstack.name,
    artifactapi_remote_helm.hashicorp_helm.name,
  ]
}
```

| Attribute      | Required | Description                               |
|----------------|----------|-------------------------------------------|
| `name`         | Yes      | Unique name (forces replacement on change)|
| `package_type` | Yes      | Package type of member remotes            |
| `description`  | No       | Human-readable description                |
| `members`      | Yes      | List of remote names to include           |
