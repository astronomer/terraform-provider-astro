// Package planmodifiers holds plan modifiers shared by the provider's resources.
package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const unknownOnUpdateDescription = "The API owns this value and may change it at any time, so it is only known after apply"

// UnknownStringOnUpdate plans a computed string as unknown whenever the resource is updated.
//
// Terraform carries a computed attribute's prior value into the plan as a known value and then
// holds the provider to it: if apply returns anything else, it fails the run with "Provider
// produced inconsistent result after apply". That contract only holds for values the provider
// controls. For a value the API owns and moves on its own schedule, the refreshed value can
// already be stale by the time apply reads the update response, and the run dies on a
// difference neither Terraform nor the provider caused. Planning the attribute as unknown says
// what is actually true: the value is settled at apply, not before.
//
// Create and destroy are left alone. The framework already plans computed attributes as unknown
// on create, and a destroy plan has nothing to modify.
func UnknownStringOnUpdate() planmodifier.String {
	return unknownStringOnUpdate{}
}

type unknownStringOnUpdate struct{}

func (m unknownStringOnUpdate) Description(_ context.Context) string {
	return unknownOnUpdateDescription
}

func (m unknownStringOnUpdate) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m unknownStringOnUpdate) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !isUpdate(req.State.Raw.IsNull(), req.Plan.Raw.IsNull()) {
		return
	}
	resp.PlanValue = types.StringUnknown()
}

// UnknownObjectOnUpdate is the object equivalent of UnknownStringOnUpdate.
func UnknownObjectOnUpdate() planmodifier.Object {
	return unknownObjectOnUpdate{}
}

type unknownObjectOnUpdate struct{}

func (m unknownObjectOnUpdate) Description(_ context.Context) string {
	return unknownOnUpdateDescription
}

func (m unknownObjectOnUpdate) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m unknownObjectOnUpdate) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !isUpdate(req.State.Raw.IsNull(), req.Plan.Raw.IsNull()) {
		return
	}
	// A null or unknown object still carries its attribute types, so this holds the object's
	// shape whether or not the deployment currently has a value for it.
	resp.PlanValue = types.ObjectUnknown(req.PlanValue.AttributeTypes(ctx))
}

// isUpdate reports whether the whole resource is being updated: a create has no prior state and
// a destroy has no plan.
func isUpdate(stateIsNull, planIsNull bool) bool {
	return !stateIsNull && !planIsNull
}
