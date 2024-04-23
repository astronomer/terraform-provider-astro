package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/utils"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
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
	Regions                    types.List   `tfsdk:"regions"`
	DefaultNodeInstance        types.Object `tfsdk:"default_node_instance"`
	NodeInstances              types.List   `tfsdk:"node_instances"`
	DefaultDatabaseInstance    types.Object `tfsdk:"default_database_instance"`
	DatabaseInstances          types.List   `tfsdk:"database_instances"`
}

func (data *ClusterOptionDataSource) ReadFromResponse(
	ctx context.Context,
	clusterOption *platform.ClusterOptions,
) diag.Diagnostics {
	data.Provider = types.StringValue(string(clusterOption.Provider))
	data.DefaultVpcSubnetRange = types.StringValue(clusterOption.DefaultVpcSubnetRange)
	data.DefaultPodSubnetRange = types.StringPointerValue(clusterOption.DefaultPodSubnetRange)
	data.DefaultServiceSubnetRange = types.StringPointerValue(clusterOption.DefaultServiceSubnetRange)
	data.DefaultServicePeeringRange = types.StringPointerValue(clusterOption.DefaultServicePeeringRange)
	data.NodeCountMin = types.Int64Value(int64(clusterOption.NodeCountMin))
	data.NodeCountMax = types.Int64Value(int64(clusterOption.NodeCountMax))
	data.NodeCountDefault = types.Int64Value(int64(clusterOption.NodeCountDefault))
	var diags diag.Diagnostics
	data.DefaultRegion, diags = RegionTypesObject(ctx, clusterOption.DefaultRegion)
	if diags.HasError() {
		return diags
	}

	data.Regions, diags = utils.ObjectList(ctx, &clusterOption.Regions, schemas.RegionAttributeTypes(), RegionTypesObject)
	if diags.HasError() {
		return diags
	}
	data.DefaultNodeInstance, diags = ProviderInstanceObject(ctx, clusterOption.DefaultNodeInstance)
	if diags.HasError() {
		return diags
	}

	data.NodeInstances, diags = utils.ObjectList(ctx, &clusterOption.NodeInstances, schemas.ProviderInstanceAttributeTypes(), ProviderInstanceObject)
	if diags.HasError() {
		return diags
	}

	data.DefaultDatabaseInstance, diags = ProviderInstanceObject(ctx, clusterOption.DefaultDatabaseInstance)
	if diags.HasError() {
		return diags
	}

	data.DatabaseInstances, diags = utils.ObjectList(ctx, &clusterOption.DatabaseInstances, schemas.ProviderInstanceAttributeTypes(), ProviderInstanceObject)
	if diags.HasError() {
		return diags
	}

	return nil
}

type Region struct {
	Name            types.String `tfsdk:"name"`
	Limited         types.Bool   `tfsdk:"limited"`
	BannedInstances types.List   `tfsdk:"banned_instances"`
}

func RegionTypesObject(
	ctx context.Context,
	regionInput platform.ProviderRegion,
) (regionOutput types.Object, diags diag.Diagnostics) {
	region := Region{
		Name: types.StringValue(regionInput.Name),
	}
	region.BannedInstances, diags = utils.StringList(regionInput.BannedInstances)
	if diags.HasError() {
		return regionOutput, diags
	}

	if regionInput.Limited != nil {
		val := types.BoolValue(*regionInput.Limited)
		region.Limited = val
	}
	return types.ObjectValueFrom(ctx, schemas.RegionAttributeTypes(), region)
}

type ProviderInstance struct {
	Name   types.String `tfsdk:"name"`
	Memory types.String `tfsdk:"memory"`
	Cpu    types.Int64  `tfsdk:"cpu"`
}

func ProviderInstanceObject(
	ctx context.Context,
	providerInstanceInput platform.ProviderInstanceType,
) (types.Object, diag.Diagnostics) {
	providerInstance := ProviderInstance{
		Name:   types.StringValue(providerInstanceInput.Name),
		Memory: types.StringValue(providerInstanceInput.Memory),
		Cpu:    types.Int64Value(int64(providerInstanceInput.Cpu)),
	}
	return types.ObjectValueFrom(ctx, schemas.ProviderInstanceAttributeTypes(), providerInstance)
}
