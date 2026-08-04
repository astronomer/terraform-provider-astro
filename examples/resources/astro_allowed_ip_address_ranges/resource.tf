# The astro_allowed_ip_address_ranges resource manages an organization's full IP access list as a
# single resource. It is authoritative: any ranges not listed here are removed on apply.
#
# The resource batches the underlying create/delete calls and automatically chunks requests that
# exceed the API's per-request limit (1000 CIDRs).

resource "astro_allowed_ip_address_ranges" "org_allow_list" {
  ip_address_ranges = [
    "203.0.113.0/24",
    "198.51.100.5/32",
  ]
}
