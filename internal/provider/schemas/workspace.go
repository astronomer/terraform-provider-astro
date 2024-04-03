package schemas

import (
	"github.com/astronomer/astronomer-terraform-provider/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func WorkspaceDataSourceSchemaAttributes() map[string]datasource.Attribute {
	return map[string]datasource.Attribute{
		"id": datasource.StringAttribute{
			MarkdownDescription: "Workspace identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasource.StringAttribute{
			MarkdownDescription: "Workspace name",
			Computed:            true,
		},
		"description": datasource.StringAttribute{
			MarkdownDescription: "Workspace description",
			Computed:            true,
		},
		"organization_name": datasource.StringAttribute{
			MarkdownDescription: "Workspace organization name",
			Computed:            true,
		},
		"cicd_enforced_default": datasource.BoolAttribute{
			MarkdownDescription: "Whether new Deployments enforce CI/CD deploys by default",
			Computed:            true,
		},
		"created_at": datasource.StringAttribute{
			MarkdownDescription: "Workspace creation timestamp",
			Computed:            true,
		},
		"updated_at": datasource.StringAttribute{
			MarkdownDescription: "Workspace last updated timestamp",
			Computed:            true,
		},
		"created_by": datasource.SingleNestedAttribute{
			MarkdownDescription: "Workspace creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchema(),
		},
		"updated_by": datasource.SingleNestedAttribute{
			MarkdownDescription: "Workspace updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchema(),
		},
	}
}

func WorkspaceResourceSchemaAttributes() map[string]resource.Attribute {
	return map[string]resource.Attribute{
		"id": resource.StringAttribute{
			MarkdownDescription: "Workspace identifier",
			Computed:            true,
			Validators:          []validator.String{validators.IsCuid()},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resource.StringAttribute{
			MarkdownDescription: "Workspace name",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtMost(50)},
		},
		"description": resource.StringAttribute{
			MarkdownDescription: "Workspace description",
			Required:            true,
		},
		"organization_name": resource.StringAttribute{
			MarkdownDescription: "Workspace organization name",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"cicd_enforced_default": resource.BoolAttribute{
			MarkdownDescription: "Whether new Deployments enforce CI/CD deploys by default",
			Required:            true,
		},
		"created_at": resource.StringAttribute{
			MarkdownDescription: "Workspace creation timestamp",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": resource.StringAttribute{
			MarkdownDescription: "Workspace last updated timestamp",
			Computed:            true,
		},
		"created_by": resource.SingleNestedAttribute{
			MarkdownDescription: "Workspace creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchema(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_by": resource.SingleNestedAttribute{
			MarkdownDescription: "Workspace updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchema(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
