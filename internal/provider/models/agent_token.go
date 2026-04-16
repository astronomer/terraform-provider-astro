package models

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AgentTokenResource defines the resource data model.
type AgentTokenResource struct {
	Id                 types.String `tfsdk:"id"`
	DeploymentId       types.String `tfsdk:"deployment_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ExpiryPeriodInDays types.Int64  `tfsdk:"expiry_period_in_days"`
	Token              types.String `tfsdk:"token"`
}

func (data *AgentTokenResource) ReadFromResponse(apiToken *iam.ApiToken, existingToken string) {
	data.Id = types.StringValue(apiToken.Id)
	data.Name = types.StringValue(apiToken.Name)
	if apiToken.Description == "" {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(apiToken.Description)
	}
	if apiToken.ExpiryPeriodInDays != nil {
		data.ExpiryPeriodInDays = types.Int64Value(int64(*apiToken.ExpiryPeriodInDays))
	} else {
		data.ExpiryPeriodInDays = types.Int64Null()
	}
	if apiToken.Token != nil && len(*apiToken.Token) > 0 {
		data.Token = types.StringValue(*apiToken.Token)
	} else if existingToken != "" {
		data.Token = types.StringValue(existingToken)
	} else {
		data.Token = types.StringNull()
	}
}
