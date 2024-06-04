package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ResourceHybridClusterWorkspaceAuthorizationSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"cluster_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the hybrid cluster to set authorizations for",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"workspace_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The IDs of the workspaces to authorize for the hybrid cluster",
			Optional:            true,
		},
	}
}
