package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClusterResource describes the resource data model.
type ClusterResource struct {
	Id                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	CloudProvider       types.String   `tfsdk:"cloud_provider"`
	DbInstanceType      types.String   `tfsdk:"db_instance_type"`
	Region              types.String   `tfsdk:"region"`
	PodSubnetRange      types.String   `tfsdk:"pod_subnet_range"`
	ServicePeeringRange types.String   `tfsdk:"service_peering_range"`
	ServiceSubnetRange  types.String   `tfsdk:"service_subnet_range"`
	VpcSubnetRange      types.String   `tfsdk:"vpc_subnet_range"`
	Metadata            types.Object   `tfsdk:"metadata"`
	Status              types.String   `tfsdk:"status"`
	CreatedAt           types.String   `tfsdk:"created_at"`
	UpdatedAt           types.String   `tfsdk:"updated_at"`
	Type                types.String   `tfsdk:"type"`
	TenantId            types.String   `tfsdk:"tenant_id"`
	ProviderAccount     types.String   `tfsdk:"provider_account"`
	NodePools           types.List     `tfsdk:"node_pools"`
	WorkspaceIds        types.List     `tfsdk:"workspace_ids"`
	IsLimited           types.Bool     `tfsdk:"is_limited"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

// ClusterDataSource describes the data source data model.
type ClusterDataSource struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	CloudProvider       types.String `tfsdk:"cloud_provider"`
	DbInstanceType      types.String `tfsdk:"db_instance_type"`
	Region              types.String `tfsdk:"region"`
	PodSubnetRange      types.String `tfsdk:"pod_subnet_range"`
	ServicePeeringRange types.String `tfsdk:"service_peering_range"`
	ServiceSubnetRange  types.String `tfsdk:"service_subnet_range"`
	VpcSubnetRange      types.String `tfsdk:"vpc_subnet_range"`
	Metadata            types.Object `tfsdk:"metadata"`
	Status              types.String `tfsdk:"status"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	Type                types.String `tfsdk:"type"`
	TenantId            types.String `tfsdk:"tenant_id"`
	ProviderAccount     types.String `tfsdk:"provider_account"`
	NodePools           types.List   `tfsdk:"node_pools"`
	WorkspaceIds        types.List   `tfsdk:"workspace_ids"`
	Tags                types.List   `tfsdk:"tags"`
	IsLimited           types.Bool   `tfsdk:"is_limited"`
}

type ClusterTag struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type NodePool struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ClusterId              types.String `tfsdk:"cluster_id"`
	CloudProvider          types.String `tfsdk:"cloud_provider"`
	MaxNodeCount           types.Int64  `tfsdk:"max_node_count"`
	NodeInstanceType       types.String `tfsdk:"node_instance_type"`
	IsDefault              types.Bool   `tfsdk:"is_default"`
	SupportedAstroMachines types.List   `tfsdk:"supported_astro_machines"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
}

func (data *ClusterResource) ReadFromResponse(
	ctx context.Context,
	cluster *platform.Cluster,
) diag.Diagnostics {
	data.Id = types.StringValue(cluster.Id)
	data.Name = types.StringValue(cluster.Name)
	data.CloudProvider = types.StringValue(string(cluster.CloudProvider))
	data.DbInstanceType = types.StringValue(cluster.DbInstanceType)
	data.Region = types.StringValue(cluster.Region)
	data.PodSubnetRange = types.StringPointerValue(cluster.PodSubnetRange)
	data.ServicePeeringRange = types.StringPointerValue(cluster.ServicePeeringRange)
	data.ServiceSubnetRange = types.StringPointerValue(cluster.ServiceSubnetRange)
	data.VpcSubnetRange = types.StringValue(cluster.VpcSubnetRange)
	var diags diag.Diagnostics
	data.Metadata, diags = ClusterMetadataTypesObject(ctx, cluster.Metadata)
	if diags.HasError() {
		return diags
	}
	data.Status = types.StringValue(string(cluster.Status))
	data.CreatedAt = types.StringValue(cluster.CreatedAt.String())
	data.UpdatedAt = types.StringValue(cluster.UpdatedAt.String())
	data.Type = types.StringValue(string(cluster.Type))
	data.TenantId = types.StringPointerValue(cluster.TenantId)
	data.ProviderAccount = types.StringPointerValue(cluster.ProviderAccount)
	data.NodePools, diags = utils.ObjectList(ctx, cluster.NodePools, schemas.NodePoolAttributeTypes(), NodePoolTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceIds, diags = utils.StringList(cluster.WorkspaceIds)
	if diags.HasError() {
		return diags
	}
	data.IsLimited = types.BoolPointerValue(cluster.IsLimited)

	return nil
}

func (data *ClusterDataSource) ReadFromResponse(
	ctx context.Context,
	cluster *platform.Cluster,
) diag.Diagnostics {
	data.Id = types.StringValue(cluster.Id)
	data.Name = types.StringValue(cluster.Name)
	data.CloudProvider = types.StringValue(string(cluster.CloudProvider))
	data.DbInstanceType = types.StringValue(cluster.DbInstanceType)
	data.Region = types.StringValue(cluster.Region)
	data.PodSubnetRange = types.StringPointerValue(cluster.PodSubnetRange)
	data.ServicePeeringRange = types.StringPointerValue(cluster.ServicePeeringRange)
	data.ServiceSubnetRange = types.StringPointerValue(cluster.ServiceSubnetRange)
	data.VpcSubnetRange = types.StringValue(cluster.VpcSubnetRange)
	var diags diag.Diagnostics
	data.Metadata, diags = ClusterMetadataTypesObject(ctx, cluster.Metadata)
	if diags.HasError() {
		return diags
	}
	data.Status = types.StringValue(string(cluster.Status))
	data.CreatedAt = types.StringValue(cluster.CreatedAt.String())
	data.UpdatedAt = types.StringValue(cluster.UpdatedAt.String())
	data.Type = types.StringValue(string(cluster.Type))
	data.TenantId = types.StringPointerValue(cluster.TenantId)
	data.ProviderAccount = types.StringPointerValue(cluster.ProviderAccount)
	data.NodePools, diags = utils.ObjectList(ctx, cluster.NodePools, schemas.NodePoolAttributeTypes(), NodePoolTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceIds, diags = utils.StringList(cluster.WorkspaceIds)
	if diags.HasError() {
		return diags
	}
	data.Tags, diags = utils.ObjectList(ctx, cluster.Tags, schemas.ClusterTagAttributeTypes(), ClusterTagTypesObject)
	if diags.HasError() {
		return diags
	}
	data.IsLimited = types.BoolPointerValue(cluster.IsLimited)

	return nil
}

func ClusterTagTypesObject(
	ctx context.Context,
	tag platform.ClusterK8sTag,
) (types.Object, diag.Diagnostics) {
	obj := ClusterTag{
		Key:   types.StringPointerValue(tag.Key),
		Value: types.StringPointerValue(tag.Value),
	}

	return types.ObjectValueFrom(ctx, schemas.ClusterTagAttributeTypes(), obj)
}

func NodePoolTypesObject(
	ctx context.Context,
	nodePool platform.NodePool,
) (types.Object, diag.Diagnostics) {
	supportedAstroMachines, diags := utils.StringList(nodePool.SupportedAstroMachines)
	if diags.HasError() {
		return types.ObjectNull(schemas.NodePoolAttributeTypes()), diags
	}
	obj := NodePool{
		Id:                     types.StringValue(nodePool.Id),
		Name:                   types.StringValue(nodePool.Name),
		ClusterId:              types.StringValue(nodePool.ClusterId),
		CloudProvider:          types.StringValue(string(nodePool.CloudProvider)),
		MaxNodeCount:           types.Int64Value(int64(nodePool.MaxNodeCount)),
		NodeInstanceType:       types.StringValue(nodePool.NodeInstanceType),
		IsDefault:              types.BoolValue(nodePool.IsDefault),
		SupportedAstroMachines: supportedAstroMachines,
		CreatedAt:              types.StringValue(nodePool.CreatedAt.String()),
		UpdatedAt:              types.StringValue(nodePool.UpdatedAt.String()),
	}

	return types.ObjectValueFrom(ctx, schemas.NodePoolAttributeTypes(), obj)
}

type ClusterMetadata struct {
	OidcIssuerUrl types.String `tfsdk:"oidc_issuer_url"`
	ExternalIps   types.List   `tfsdk:"external_ips"`
}

func ClusterMetadataTypesObject(
	ctx context.Context,
	metadata *platform.ClusterMetadata,
) (types.Object, diag.Diagnostics) {
	if metadata != nil {
		externalIps, diags := utils.StringList(metadata.ExternalIPs)
		if diags.HasError() {
			return types.ObjectNull(schemas.ClusterMetadataAttributeTypes()), diags
		}
		obj := ClusterMetadata{
			OidcIssuerUrl: types.StringPointerValue(metadata.OidcIssuerUrl),
			ExternalIps:   externalIps,
		}
		return types.ObjectValueFrom(ctx, schemas.ClusterMetadataAttributeTypes(), obj)
	}
	return types.ObjectNull(schemas.ClusterMetadataAttributeTypes()), nil
}
