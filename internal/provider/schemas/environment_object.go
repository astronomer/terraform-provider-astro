package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func EnvironmentObjectSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Environment Object identifier",
			Computed:            true,
		},
		"object_key": schema.StringAttribute{
			MarkdownDescription: "Key of the environment object",
			Required:            true,
		},
		"object_type": schema.StringAttribute{
			MarkdownDescription: "Type of the environment object (e.g., CONNECTION, AIRFLOW_VARIABLE)",
			Required:            true,
		},
		"scope": schema.StringAttribute{
			MarkdownDescription: "Scope of the environment object (e.g., WORKSPACE, DEPLOYMENT)",
			Required:            true,
		},
		"scope_entity_id": schema.StringAttribute{
			MarkdownDescription: "ID of the entity within the scope",
			Required:            true,
		},
		// Add additional attributes as needed
	}
}
