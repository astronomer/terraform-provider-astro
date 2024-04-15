package resources

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/samber/lo"
)

// RequestScalingSpec converts a Terraform object to a platform.DeploymentScalingSpecRequest to be used in create and update requests
func RequestScalingSpec(ctx context.Context, scalingSpecObj types.Object) (*platform.DeploymentScalingSpecRequest, diag.Diagnostics) {
	if scalingSpecObj.IsNull() {
		return nil, nil
	}

	var scalingSpec models.DeploymentScalingSpec
	diags := scalingSpecObj.As(ctx, &scalingSpec, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    false,
		UnhandledUnknownAsEmpty: false,
	})
	if diags.HasError() {
		return nil, diags
	}
	platformScalingSpec := &platform.DeploymentScalingSpecRequest{
		HibernationSpec: &platform.DeploymentHibernationSpecRequest{
			Override: &platform.DeploymentHibernationOverrideRequest{
				IsHibernating: scalingSpec.HibernationSpec.Override.IsHibernating.ValueBoolPointer(),
				OverrideUntil: scalingSpec.HibernationSpec.Override.OverrideUntil.ValueStringPointer(),
			},
		},
	}
	if len(scalingSpec.HibernationSpec.Schedules) > 0 {
		schedules := lo.Map(scalingSpec.HibernationSpec.Schedules, func(schedule models.HibernationSchedule, _ int) platform.DeploymentHibernationSchedule {
			return platform.DeploymentHibernationSchedule{
				Description:     schedule.Description.ValueStringPointer(),
				HibernateAtCron: schedule.HibernateAtCron.ValueString(),
				IsEnabled:       schedule.IsEnabled.ValueBool(),
				WakeAtCron:      schedule.WakeAtCron.ValueString(),
			}
		})
		platformScalingSpec.HibernationSpec.Schedules = &schedules
	}
	return platformScalingSpec, nil
}

// RequestHostedWorkerQueues converts a Terraform list to a list of platform.WorkerQueueRequest to be used in create and update requests
func RequestHostedWorkerQueues(ctx context.Context, workerQueuesObjList types.List) (*[]platform.WorkerQueueRequest, diag.Diagnostics) {
	if len(workerQueuesObjList.Elements()) == 0 {
		return nil, nil
	}

	var workerQueues []models.HostedWorkerQueue
	diags := workerQueuesObjList.ElementsAs(ctx, &workerQueues, false)
	if diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.HostedWorkerQueue, _ int) platform.WorkerQueueRequest {
		return platform.WorkerQueueRequest{
			AstroMachine:      platform.WorkerQueueRequestAstroMachine(workerQueue.AstroMachine.ValueString()),
			IsDefault:         workerQueue.IsDefault.ValueBool(),
			MaxWorkerCount:    int(workerQueue.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:    int(workerQueue.MinWorkerCount.ValueInt64()),
			Name:              workerQueue.Name.ValueString(),
			WorkerConcurrency: int(workerQueue.WorkerConcurrency.ValueInt64()),
		}
	})
	return &platformWorkerQueues, nil
}

// RequestDeploymentEnvironmentVariables converts a Terraform list to a list of platform.DeploymentEnvironmentVariableRequest to be used in create and update requests
func RequestDeploymentEnvironmentVariables(ctx context.Context, environmentVariablesObjList types.List) ([]platform.DeploymentEnvironmentVariableRequest, diag.Diagnostics) {
	if len(environmentVariablesObjList.Elements()) == 0 {
		return []platform.DeploymentEnvironmentVariableRequest{}, nil
	}

	var envVars []models.DeploymentEnvironmentVariable
	diags := environmentVariablesObjList.ElementsAs(ctx, &envVars, false)
	if diags.HasError() {
		return nil, diags
	}
	platformEnvVars := lo.Map(envVars, func(envVar models.DeploymentEnvironmentVariable, _ int) platform.DeploymentEnvironmentVariableRequest {
		return platform.DeploymentEnvironmentVariableRequest{
			IsSecret: envVar.IsSecret.ValueBool(),
			Key:      envVar.Key.ValueString(),
			Value:    envVar.Value.ValueStringPointer(),
		}
	})
	return platformEnvVars, nil
}
