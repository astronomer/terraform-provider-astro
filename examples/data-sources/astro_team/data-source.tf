data "astro_team" "example" {
  id = "clwbclrc100bl01ozjj5s4jmq"
}

# Output the team value using terraform apply
output "team" {
  value = data.astro_team.example
}