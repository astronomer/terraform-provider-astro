package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WorkspaceRoleAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"workspace_id": types.StringType,
		"role":         types.StringType,
	}
}

func ResourceWorkspaceRoleSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"workspace_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the workspace to assign the role to",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "The role to assign to the workspace",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(iam.WORKSPACEOWNER),
					string(iam.WORKSPACEMEMBER),
					string(iam.WORKSPACEACCESSOR),
					string(iam.WORKSPACEOPERATOR),
					string(iam.WORKSPACEAUTHOR),
				),
			},
		},
	}
}

func DataSourceWorkspaceRoleSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"workspace_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the workspace the role is assigned to",
			Computed:            true,
		},
		"role": datasourceSchema.StringAttribute{
			MarkdownDescription: "The role assigned to the workspace",
			Computed:            true,
		},
	}
}

func DeploymentRoleAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"deployment_id": types.StringType,
		"role":          types.StringType,
	}
}

func ResourceDeploymentRoleSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the deployment to assign the role to",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "The role to assign to the deployment",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}

func DataSourceDeploymentRoleSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the deployment the role is assigned to",
			Computed:            true,
		},
		"role": datasourceSchema.StringAttribute{
			MarkdownDescription: "The role assigned to the deployment",
			Computed:            true,
		},
	}
}

func DagRoleAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dag_id":        types.StringType,
		"deployment_id": types.StringType,
		"role":          types.StringType,
		"tag":           types.StringType,
	}
}

func ResourceDagRoleSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"dag_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The DAG ID. Required if tag is not specified.",
			Optional:            true,
		},
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The Deployment ID containing the DAG.",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "The DAG role (DAG_VIEWER, DAG_AUTHOR, or custom DAG role).",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"tag": resourceSchema.StringAttribute{
			MarkdownDescription: "The DAG tag. Required if dag_id is not specified.",
			Optional:            true,
		},
	}
}

func DataSourceDagRoleSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"dag_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The DAG ID",
			Computed:            true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The Deployment ID containing the DAG",
			Computed:            true,
		},
		"role": datasourceSchema.StringAttribute{
			MarkdownDescription: "The DAG role",
			Computed:            true,
		},
		"tag": datasourceSchema.StringAttribute{
			MarkdownDescription: "The DAG tag",
			Computed:            true,
		},
	}
}

func ApiTokenRoleAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"entity_id":     types.StringType,
		"entity_type":   types.StringType,
		"role":          types.StringType,
		"deployment_id": types.StringType,
	}
}

func ResourceApiTokenRoleSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the entity to assign the role to. For DAG entity type, this is the dag_id. For TAG entity type, this is the tag value.",
			Required:            true,
		},
		"entity_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of entity to assign the role to",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(iam.ApiTokenRoleEntityTypeORGANIZATION),
					string(iam.ApiTokenRoleEntityTypeWORKSPACE),
					string(iam.ApiTokenRoleEntityTypeDEPLOYMENT),
					string(iam.ApiTokenRoleEntityTypeDAG),
					string(iam.ApiTokenRoleEntityTypeTAG),
				),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "The role to assign to the entity",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The Deployment ID. Required for DAG and TAG entity types.",
			Optional:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
	}
}

func DataSourceApiTokenRoleSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the entity to assign the role to",
			Computed:            true,
		},
		"entity_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of entity to assign the role to",
			Computed:            true,
		},
		"role": datasourceSchema.StringAttribute{
			MarkdownDescription: "The role to assign to the entity",
			Computed:            true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The Deployment ID for DAG and TAG entity types",
			Computed:            true,
		},
	}
}
