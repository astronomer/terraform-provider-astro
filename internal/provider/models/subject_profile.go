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
	case iam.BasicSubjectProfile:
		bspPtr = iamSubjectProfileToPlatform(&v)
	case *iam.BasicSubjectProfile:
		bspPtr = iamSubjectProfileToPlatform(v)
	case platform_v1.BasicSubjectProfile:
		bspPtr = platformV1SubjectProfileToPlatform(&v)
	case *platform_v1.BasicSubjectProfile:
		bspPtr = platformV1SubjectProfileToPlatform(v)
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

	// The API omits fields like a Team's updatedBy, which arrive here as typed nils
	if bspPtr == nil {
		return types.ObjectNull(schemas.SubjectProfileAttributeTypes()), nil
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

func iamSubjectProfileToPlatform(bsp *iam.BasicSubjectProfile) *platform.BasicSubjectProfile {
	if bsp == nil {
		return nil
	}
	return &platform.BasicSubjectProfile{
		ApiTokenName: bsp.ApiTokenName,
		AvatarUrl:    bsp.AvatarUrl,
		FullName:     bsp.FullName,
		Id:           bsp.Id,
		SubjectType:  (*platform.BasicSubjectProfileSubjectType)(bsp.SubjectType),
		Username:     bsp.Username,
	}
}

// platform_v1 has the same field shape but distinct types.
func platformV1SubjectProfileToPlatform(bsp *platform_v1.BasicSubjectProfile) *platform.BasicSubjectProfile {
	if bsp == nil {
		return nil
	}
	return &platform.BasicSubjectProfile{
		ApiTokenName: bsp.ApiTokenName,
		AvatarUrl:    bsp.AvatarUrl,
		FullName:     bsp.FullName,
		Id:           bsp.Id,
		SubjectType:  (*platform.BasicSubjectProfileSubjectType)(bsp.SubjectType),
		Username:     bsp.Username,
	}
}
