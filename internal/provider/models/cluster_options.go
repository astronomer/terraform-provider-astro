package models

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClusterOptionsDataSource describes the data source data model.
type ClusterOptionsDataSource struct {
	ClusterOptions types.List   `tfsdk:"cluster_options"`
	Type           types.String `tfsdk:"type"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
}

func (data *ClusterOptionsDataSource) ReadFromResponse(
	ctx context.Context,
	clusterOptions []platform.ClusterOptions,
) diag.Diagnostics {
	if len(clusterOptions) == 0 {
		types.ListNull(types.ObjectType{AttrTypes: schemas.ClusterOptionsElementAttributeTypes()})
	}

	values := make([]attr.Value, len(clusterOptions))
	for i, clusterOption := range clusterOptions {
		var data ClusterOptionDataSource
		diags := data.ReadFromResponse(ctx, &clusterOption)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.ClusterOptionsElementAttributeTypes(), data)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.ClusterOptions, diags = types.ListValue(types.ObjectType{AttrTypes: schemas.ClusterOptionsElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}
	return nil
}

// ClusterOptionsDataSource describes the data source data model.
type ClusterOptionDataSource struct {
	Provider                   types.String `tfsdk:"provider"`
	DefaultVpcSubnetRange      types.String `tfsdk:"default_vpc_subnet_range"`
	DefaultPodSubnetRange      types.String `tfsdk:"default_pod_subnet_range"`
	DefaultServiceSubnetRange  types.String `tfsdk:"default_service_subnet_range"`
	DefaultServicePeeringRange types.String `tfsdk:"default_service_peering_range"`
	NodeCountMin               types.Int64  `tfsdk:"node_count_min"`
	NodeCountMax               types.Int64  `tfsdk:"node_count_max"`
	NodeCountDefault           types.Int64  `tfsdk:"node_count_default"`
	DefaultRegion              types.Object `tfsdk:"default_region"`
}

func (data *ClusterOptionDataSource) ReadFromResponse(
	ctx context.Context,
	clusterOption *platform.ClusterOptions,
) diag.Diagnostics {
	data.Provider = types.StringValue(string(clusterOption.Provider))
	data.DefaultVpcSubnetRange = types.StringValue(clusterOption.DefaultVpcSubnetRange)
	if clusterOption.DefaultPodSubnetRange != nil {
		data.DefaultPodSubnetRange = types.StringValue(*clusterOption.DefaultPodSubnetRange)
	}
	if clusterOption.DefaultServiceSubnetRange != nil {
		data.DefaultServiceSubnetRange = types.StringValue(*clusterOption.DefaultServiceSubnetRange)
	}
	if clusterOption.DefaultServicePeeringRange != nil {
		data.DefaultServicePeeringRange = types.StringValue(*clusterOption.DefaultServicePeeringRange)
	}

	data.NodeCountMin = types.Int64Value(int64(clusterOption.NodeCountMin))
	data.NodeCountMax = types.Int64Value(int64(clusterOption.NodeCountMax))
	data.NodeCountDefault = types.Int64Value(int64(clusterOption.NodeCountDefault))
	var diags diag.Diagnostics
	data.DefaultRegion, diags = DefaultRegionTypesObject(ctx, clusterOption.DefaultRegion)
	if diags.HasError() {
		return diags
	}

	return nil
}

type DefaultRegion struct {
	Name    types.String `tfsdk:"name"`
	Limited types.Bool   `tfsdk:"limited"`
	//BannedInstances types.List   `tfsdk:"banned_instances"`
}

func DefaultRegionTypesObject(
	ctx context.Context,
	defaultRegionInput platform.ProviderRegion,
) (defaultRegionOutput types.Object, diags diag.Diagnostics) {
	defaultRegion := DefaultRegion{
		Name: types.StringValue(defaultRegionInput.Name),
	}
	//defaultRegion.BannedInstances, diags = utils.StringList(defaultRegionInput.BannedInstances)
	//if diags.HasError() {
	//	return defaultRegionOutput, diags
	//}

	if defaultRegionInput.Limited != nil {
		val := types.BoolValue(*defaultRegionInput.Limited)
		defaultRegion.Limited = val
	}
	return types.ObjectValueFrom(ctx, schemas.DefaultRegionAttributeTypes(), defaultRegion)
}
