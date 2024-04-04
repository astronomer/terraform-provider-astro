package models

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/iam"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type SubjectProfile struct {
	Id           types.String `tfsdk:"id"`
	SubjectType  types.String `tfsdk:"subject_type"`
	Username     types.String `tfsdk:"username"`
	FullName     types.String `tfsdk:"full_name"`
	AvatarUrl    types.String `tfsdk:"avatar_url"`
	ApiTokenName types.String `tfsdk:"api_token_name"`
}

var SubjectProfileTF map[string]attr.Type = map[string]attr.Type{
	"id":             types.StringType,
	"subject_type":   types.StringType,
	"username":       types.StringType,
	"full_name":      types.StringType,
	"avatar_url":     types.StringType,
	"api_token_name": types.StringType,
}

func SubjectProfileTypesObject(
	ctx context.Context,
	basicSubjectProfile any,
) (types.Object, diag.Diagnostics) {
	// Check that the type passed in is a platform.BasicSubjectProfile or iam.BasicSubjectProfile
	bsp, ok := basicSubjectProfile.(*platform.BasicSubjectProfile)
	if !ok {
		iamBsp, ok := basicSubjectProfile.(*iam.BasicSubjectProfile)
		if !ok {
			tflog.Error(
				ctx,
				"Unexpected type passed into subject profile",
				map[string]interface{}{"value": basicSubjectProfile},
			)
			return types.Object{}, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Internal Error",
					"SubjectProfileTypesObject expects a BasicSubjectProfile type but did not receive one",
				),
			}
		}
		// Convert the iam.BasicSubjectProfile to a platform.BasicSubjectProfile for simplicity
		bsp = &platform.BasicSubjectProfile{
			ApiTokenName: iamBsp.ApiTokenName,
			AvatarUrl:    iamBsp.AvatarUrl,
			FullName:     iamBsp.FullName,
			Id:           iamBsp.Id,
			SubjectType:  (*platform.BasicSubjectProfileSubjectType)(iamBsp.SubjectType),
			Username:     iamBsp.Username,
		}
	}

	subjectProfile := SubjectProfile{
		Id: types.StringValue(bsp.Id),
	}

	if bsp.SubjectType != nil {
		subjectProfile.SubjectType = types.StringValue(string(*bsp.SubjectType))
		if *bsp.SubjectType == platform.USER {
			if bsp.Username != nil {
				subjectProfile.Username = types.StringValue(*bsp.Username)
			} else {
				subjectProfile.Username = types.StringUnknown()
			}
			if bsp.FullName != nil {
				subjectProfile.FullName = types.StringValue(*bsp.FullName)
			} else {
				subjectProfile.FullName = types.StringUnknown()
			}
			if bsp.AvatarUrl != nil {
				subjectProfile.AvatarUrl = types.StringValue(*bsp.AvatarUrl)
			} else {
				subjectProfile.AvatarUrl = types.StringUnknown()
			}
		} else {
			if bsp.ApiTokenName != nil {
				subjectProfile.ApiTokenName = types.StringValue(*bsp.ApiTokenName)
			} else {
				subjectProfile.ApiTokenName = types.StringUnknown()
			}
		}
	}
	return types.ObjectValueFrom(ctx, SubjectProfileTF, subjectProfile)
}
