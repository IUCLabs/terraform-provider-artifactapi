# Ansible remote proxies the Galaxy v3 collections API.
# The provider knows JSON metadata is mutable; collection tarballs are immutable.
resource "artifactapi_remote_ansible" "galaxy" {
  name        = "galaxy"
  base_url    = "https://galaxy.ansible.com"
  description = "Ansible Galaxy"

  immutable_ttl = 0
  mutable_ttl   = 300
}
