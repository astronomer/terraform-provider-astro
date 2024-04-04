package schemas

import (
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DataSourceSubjectProfileSchema() map[string]datasource.Attribute {
	return map[string]datasource.Attribute{
		"id": datasource.StringAttribute{
			Computed: true,
		},
		"subject_type": datasource.StringAttribute{
			Computed: true,
		},
		"username": datasource.StringAttribute{
			Computed: true,
		},
		"full_name": datasource.StringAttribute{
			Computed: true,
		},
		"avatar_url": datasource.StringAttribute{
			Computed: true,
		},
		"api_token_name": datasource.StringAttribute{
			Computed: true,
		},
	}
}

func ResourceSubjectProfileSchema() map[string]resource.Attribute {
	return map[string]resource.Attribute{
		"id": resource.StringAttribute{
			Computed: true,
		},
		"subject_type": resource.StringAttribute{
			Computed: true,
		},
		"username": resource.StringAttribute{
			Computed: true,
		},
		"full_name": resource.StringAttribute{
			Computed: true,
		},
		"avatar_url": resource.StringAttribute{
			Computed: true,
		},
		"api_token_name": resource.StringAttribute{
			Computed: true,
		},
	}
}
