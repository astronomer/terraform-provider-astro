package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func RemoteExecutionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
		"allowed_ip_address_ranges": types.SetType{
			ElemType: types.StringType,
		},
		"remote_api_url":       types.StringType,
		"task_log_bucket":      types.StringType,
		"task_log_url_pattern": types.StringType,
	}
}

func RemoteExecutionDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"enabled": datasourceSchema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether remote execution is enabled",
		},
		"allowed_ip_address_ranges": datasourceSchema.SetAttribute{
			Computed:            true,
			MarkdownDescription: "The allowed IP address ranges for remote execution",
			ElementType:         types.StringType,
		},
		"remote_api_url": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The URL for the remote API",
		},
		"task_log_bucket": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The bucket for task logs",
		},
		"task_log_url_pattern": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The URL pattern for task logs",
		},
	}
}

func RemoteExecutionResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether remote execution is enabled",
			Required:            true,
			Validators: []validator.Bool{
				// An explicit `false` is not allowed,
				// as the API removes it from the response if set to `false`,
				// which causes issues with the provider.
				boolvalidator.Equals(true),
			},
		},
		"allowed_ip_address_ranges": resourceSchema.SetAttribute{
			MarkdownDescription: "The allowed IP address ranges for remote execution",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"remote_api_url": resourceSchema.StringAttribute{
			MarkdownDescription: "The URL for the remote API",
			Computed:            true,
		},
		"task_log_bucket": resourceSchema.StringAttribute{
			MarkdownDescription: "The bucket for task logs",
			Optional:            true,
		},
		"task_log_url_pattern": resourceSchema.StringAttribute{
			MarkdownDescription: "The URL pattern for task logs",
			Optional:            true,
		},
	}
}
