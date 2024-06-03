resource "astro_workspace" "example" {
  name                  = "my-workspace"
  description           = "my first workspace"
  cicd_enforced_default = true
}

// Import an existing workspace
import = {
  id = "clozc036j01to01jrlgvu798d" // ID of the existing workspace
  to = astro_workspace.imported_workspace
}
resource "astro_workspace" "imported_workspace" {
  name                  = "import me"
  description           = "an existing workspace"
  cicd_enforced_default = true
}