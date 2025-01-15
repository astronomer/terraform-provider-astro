terraform {
  required_providers {
    astro = {
      source = "astronomer/astro"
    }
  }
}

provider "astro" {
  organization_id = "clsaoc8id051901jsmvivh82z"
}

# Add resources and data sources here for testing