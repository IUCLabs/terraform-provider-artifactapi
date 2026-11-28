# Import by name. The private key is never returned by the server, so the
# next apply rewrites it from configuration.
terraform import artifactapi_github_app.acme acme
