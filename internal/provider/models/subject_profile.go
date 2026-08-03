package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
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

func SubjectProfileTypesObject(
	ctx context.Context,
	basicSubjectProfile any,
) (types.Object, diag.Diagnostics) {
	// Attempt to convert basicSubjectProfile to *platform.BasicSubjectProfile
	// Our API client returns a BasicSubjectProfile, but we are unsure if it is a pointer and which package it is from
	var bspPtr *platform.BasicSubjectProfile

	switch v := basicSubjectProfile.(type) {
	case platform.BasicSubjectProfile:
		bspPtr = &v
	case *platform.BasicSubjectProfile:
		bspPtr = v
	case iam.BasicSubjectProfile, *iam.BasicSubjectProfile:
		var iamBsp *iam.BasicSubjectProfile
		if nonPtr, ok := v.(iam.BasicSubjectProfile); ok {
			iamBsp = &nonPtr
		} else {
			iamBsp = v.(*iam.BasicSubjectProfile)
		}

		bspPtr = &platform.BasicSubjectProfile{
			ApiTokenName: iamBsp.ApiTokenName,
			AvatarUrl:    iamBsp.AvatarUrl,
			FullName:     iamBsp.FullName,
			Id:           iamBsp.Id,
			SubjectType:  (*platform.BasicSubjectProfileSubjectType)(iamBsp.SubjectType),
			Username:     iamBsp.Username,
		}
	case platform_v1.BasicSubjectProfile, *platform_v1.BasicSubjectProfile:
		// platform_v1 (unified public API) has the same field shape but distinct
		// types — convert to *platform.BasicSubjectProfile so downstream code stays
		// unchanged.
		var v1Bsp *platform_v1.BasicSubjectProfile
		if nonPtr, ok := v.(platform_v1.BasicSubjectProfile); ok {
			v1Bsp = &nonPtr
		} else {
			v1Bsp = v.(*platform_v1.BasicSubjectProfile)
		}

		bspPtr = &platform.BasicSubjectProfile{
			ApiTokenName: v1Bsp.ApiTokenName,
			AvatarUrl:    v1Bsp.AvatarUrl,
			FullName:     v1Bsp.FullName,
			Id:           v1Bsp.Id,
			SubjectType:  (*platform.BasicSubjectProfileSubjectType)(v1Bsp.SubjectType),
			Username:     v1Bsp.Username,
		}
	default:
		// Log error and return if none of the types match
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

	subjectProfile := SubjectProfile{
		Id:           types.StringValue(bspPtr.Id),
		SubjectType:  types.StringPointerValue((*string)(bspPtr.SubjectType)),
		Username:     types.StringPointerValue(bspPtr.Username),
		FullName:     types.StringPointerValue(bspPtr.FullName),
		AvatarUrl:    types.StringPointerValue(bspPtr.AvatarUrl),
		ApiTokenName: types.StringPointerValue(bspPtr.ApiTokenName),
	}

	return types.ObjectValueFrom(ctx, schemas.SubjectProfileAttributeTypes(), subjectProfile)
}
