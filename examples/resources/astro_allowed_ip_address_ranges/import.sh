# The IP access list is a singleton per organization. Import it using the organization ID; the
# provider reads the current set of ranges into ip_address_ranges. Import (rather than re-declaring)
# to adopt an access list that already exists, otherwise the first apply conflicts with the existing
# ranges.
terraform import astro_allowed_ip_address_ranges.org_allow_list <organization_id>
