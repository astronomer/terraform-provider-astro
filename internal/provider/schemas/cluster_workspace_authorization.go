package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ResourceClusterWorkspaceAuthorizationSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"cluster_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the cluster to set authorizations for.",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"workspace_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "The IDs of the workspaces to authorize for the cluster.",
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
	}
}
