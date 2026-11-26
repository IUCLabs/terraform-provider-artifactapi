# Global server settings; there is one per server. Unset attributes
# use the server's environment defaults, and destroying the resource
# reverts every setting to them.
resource "artifactapi_settings" "this" {
  s3_health_up_threshold   = 2
  s3_health_down_threshold = 5
  s3_health_interval_ms    = 5000
  s3_health_timeout_ms     = 2000
}
