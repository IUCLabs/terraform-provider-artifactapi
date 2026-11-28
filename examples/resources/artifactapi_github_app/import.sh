# Import by name. The server never returns the private key: with private_key
# the next apply rewrites it; with private_key_wo it is resent only if its
# fingerprint differs from the server's or private_key_wo_version is set.
terraform import artifactapi_github_app.acme acme
