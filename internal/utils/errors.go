package utils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func DataSourceApiClientConfigureError(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	tflog.Error(
		ctx,
		"unexpected data source configure type",
		map[string]interface{}{"type": fmt.Sprintf("%T", req.ProviderData)},
	)
	resp.Diagnostics.AddError(
		"Unexpected Data Source Configure Type",
		fmt.Sprintf(
			"Expected apiClientsModel, got: %T. Please report this issue to the provider developers.",
			req.ProviderData,
		),
	)
	return
}

func ResourceApiClientConfigureError(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	tflog.Error(
		ctx,
		"unexpected resource configure type",
		map[string]interface{}{"type": fmt.Sprintf("%T", req.ProviderData)},
	)
	resp.Diagnostics.AddError(
		"Unexpected Resource Configure Type",
		fmt.Sprintf(
			"Expected apiClientsModel, got: %T. Please report this issue to the provider developers.",
			req.ProviderData,
		),
	)
	return
}
