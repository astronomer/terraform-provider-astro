package schemas

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
)

func AgentTokenResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Agent Token identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the deployment this agent token belongs to",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Agent Token name",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Agent Token description",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expiry_period_in_days": resourceSchema.Int64Attribute{
			MarkdownDescription: "Agent Token expiry period in days. If not set, the token will not expire.",
			Optional:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"token": resourceSchema.StringAttribute{
			MarkdownDescription: "Agent Token value. Warning: This value will be saved in plaintext in the terraform state file.",
			Computed:            true,
			Sensitive:           true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
