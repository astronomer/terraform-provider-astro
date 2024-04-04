package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var SubjectProfileTF = map[string]attr.Type{
	"id":             types.StringType,
	"subject_type":   types.StringType,
	"username":       types.StringType,
	"full_name":      types.StringType,
	"avatar_url":     types.StringType,
	"api_token_name": types.StringType,
}

func DataSourceSubjectProfileSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"subject_type": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"username": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"full_name": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"avatar_url": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"api_token_name": datasourceSchema.StringAttribute{
			Computed: true,
		},
	}
}

func ResourceSubjectProfileSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			Computed: true,
		},
		"subject_type": resourceSchema.StringAttribute{
			Computed: true,
		},
		"username": resourceSchema.StringAttribute{
			Computed: true,
		},
		"full_name": resourceSchema.StringAttribute{
			Computed: true,
		},
		"avatar_url": resourceSchema.StringAttribute{
			Computed: true,
		},
		"api_token_name": resourceSchema.StringAttribute{
			Computed: true,
		},
	}
}
