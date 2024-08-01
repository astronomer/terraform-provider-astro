resource "astro_user_roles" "organization_role_only" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_OWNER"
}

resource "astro_user_roles" "workspace_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clx42sxw501gl01o0gjenthnh"
      role         = "WORKSPACE_MEMBER"
    }
  ]
}

resource "astro_user_roles" "deployment_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  deployment_roles = [
    {
      deployment_id = "clyn6kxud003x01mtxmccegnh"
      role          = "DEPLOYMENT_ADMIN"
    }
  ]
}

resource "astro_user_roles" "all_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clx42sxw501gl01o0gjenthnh"
      role         = "WORKSPACE_OWNER"
    },
    {
      workspace_id = "clzafte7z006001lhkey6qzzb"
      role         = "WORKSPACE_MEMBER"
    }
  ]
  deployment_roles = [
    {
      deployment_id = "clyn6kxud003x01mtxmccegnh"
      role          = "my custom role"
    }
  ]
}