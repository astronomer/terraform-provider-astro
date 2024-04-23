package models

import (
	"context"
	"strconv"

	"github.com/astronomer/terraform-provider-astro/internal/utils"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeploymentOptions struct {
	Executors               types.List   `tfsdk:"executors"`
	ResourceQuotas          types.Object `tfsdk:"resource_quotas"`
	RuntimeReleases         types.List   `tfsdk:"runtime_releases"`
	SchedulerMachines       types.List   `tfsdk:"scheduler_machines"`
	WorkerMachines          types.List   `tfsdk:"worker_machines"`
	WorkerQueues            types.Object `tfsdk:"worker_queues"`
	WorkloadIdentityOptions types.List   `tfsdk:"workload_identity_options"`

	// Query params
	DeploymentId   types.String `tfsdk:"deployment_id"`
	DeploymentType types.String `tfsdk:"deployment_type"`
	Executor       types.String `tfsdk:"executor"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
}

type ResourceQuotaOptions struct {
	ResourceQuota  types.Object `tfsdk:"resource_quota"`
	DefaultPodSize types.Object `tfsdk:"default_pod_size"`
}

type ResourceOption struct {
	Cpu    types.Object `tfsdk:"cpu"`
	Memory types.Object `tfsdk:"memory"`
}

type ResourceRange struct {
	Floor   types.String `tfsdk:"floor"`
	Default types.String `tfsdk:"default"`
	Ceiling types.String `tfsdk:"ceiling"`
}

type RuntimeRelease struct {
	Version                  types.String `tfsdk:"version"`
	AirflowVersion           types.String `tfsdk:"airflow_version"`
	ReleaseDate              types.String `tfsdk:"release_date"`
	AirflowDatabaseMigration types.Bool   `tfsdk:"airflow_database_migration"`
	StellarDatabaseMigration types.Bool   `tfsdk:"stellar_database_migration"`
	Channel                  types.String `tfsdk:"channel"`
}

type MachineSpec struct {
	Cpu              types.String `tfsdk:"cpu"`
	Memory           types.String `tfsdk:"memory"`
	EphemeralStorage types.String `tfsdk:"ephemeral_storage"`
	Concurrency      types.String `tfsdk:"concurrency"`
}

type SchedulerMachine struct {
	Name types.String `tfsdk:"name"`
	Spec types.Object `tfsdk:"spec"`
}

type WorkerMachine struct {
	Name        types.String `tfsdk:"name"`
	Spec        types.Object `tfsdk:"spec"`
	Concurrency types.Object `tfsdk:"concurrency"`
}

type WorkerQueueOptions struct {
	MinWorkers        types.Object `tfsdk:"min_workers"`
	MaxWorkers        types.Object `tfsdk:"max_workers"`
	WorkerConcurrency types.Object `tfsdk:"worker_concurrency"`
}

type WorkloadIdentityOption struct {
	Role  types.String `tfsdk:"role"`
	Label types.String `tfsdk:"label"`
}

func (data *DeploymentOptions) ReadFromResponse(
	ctx context.Context,
	options *platform.DeploymentOptions,
) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Executors, diags = utils.StringList(&options.Executors)
	if diags.HasError() {
		return diags
	}
	data.ResourceQuotas, diags = ResourceQuotaOptionsObject(ctx, options.ResourceQuotas)
	if diags.HasError() {
		return diags
	}
	data.RuntimeReleases, diags = utils.ObjectList(ctx, &options.RuntimeReleases, schemas.RuntimeReleaseAttributeTypes(), RuntimeReleaseTypesObject)
	if diags.HasError() {
		return diags
	}
	data.SchedulerMachines, diags = utils.ObjectList(ctx, &options.SchedulerMachines, schemas.SchedulerMachineAttributeTypes(), SchedulerMachineTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkerMachines, diags = utils.ObjectList(ctx, &options.WorkerMachines, schemas.WorkerMachineAttributeTypes(), WorkerMachineTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkerQueues, diags = WorkerQueueOptionsTypesObject(ctx, options.WorkerQueues)
	if diags.HasError() {
		return diags
	}
	data.WorkloadIdentityOptions, diags = utils.ObjectList(ctx, options.WorkloadIdentityOptions, schemas.WorkloadIdentityOptionsAttributeTypes(), WorkloadIdentityOptionTypesObject)
	if diags.HasError() {
		return diags
	}

	return nil
}

func ResourceRangeTypesObject(
	ctx context.Context,
	resourceRange platform.ResourceRange,
) (types.Object, diag.Diagnostics) {
	obj := ResourceRange{
		Floor:   types.StringValue(resourceRange.Floor),
		Default: types.StringValue(resourceRange.Default),
		Ceiling: types.StringValue(resourceRange.Ceiling),
	}

	return types.ObjectValueFrom(ctx, schemas.ResourceRangeAttributeTypes(), obj)
}

func RangeTypesObject(
	ctx context.Context,
	range_ platform.Range,
) (types.Object, diag.Diagnostics) {
	floor := strconv.FormatFloat(float64(range_.Floor), 'f', -1, 64)
	default_ := strconv.FormatFloat(float64(range_.Default), 'f', -1, 64)
	ceiling := strconv.FormatFloat(float64(range_.Ceiling), 'f', -1, 64)
	obj := ResourceRange{
		Floor:   types.StringValue(floor),
		Default: types.StringValue(default_),
		Ceiling: types.StringValue(ceiling),
	}

	return types.ObjectValueFrom(ctx, schemas.ResourceRangeAttributeTypes(), obj)
}

func ResourceOptionTypesObject(
	ctx context.Context,
	resourceOption platform.ResourceOption,
) (types.Object, diag.Diagnostics) {
	cpu, diags := ResourceRangeTypesObject(ctx, resourceOption.Cpu)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceOptionAttributeTypes()), diags
	}
	memory, diags := ResourceRangeTypesObject(ctx, resourceOption.Memory)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceOptionAttributeTypes()), diags
	}
	obj := ResourceOption{
		Cpu:    cpu,
		Memory: memory,
	}
	return types.ObjectValueFrom(ctx, schemas.ResourceOptionAttributeTypes(), obj)
}

func ResourceQuotaOptionsObject(
	ctx context.Context,
	resourceQuotaOptions platform.ResourceQuotaOptions,
) (types.Object, diag.Diagnostics) {
	resourceQuota, diags := ResourceOptionTypesObject(ctx, resourceQuotaOptions.ResourceQuota)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceQuotaOptionsAttributeTypes()), diags
	}
	defaultPodSize, diags := ResourceOptionTypesObject(ctx, resourceQuotaOptions.DefaultPodSize)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceQuotaOptionsAttributeTypes()), diags
	}
	obj := ResourceQuotaOptions{
		ResourceQuota:  resourceQuota,
		DefaultPodSize: defaultPodSize,
	}

	return types.ObjectValueFrom(ctx, schemas.ResourceQuotaOptionsAttributeTypes(), obj)
}

func RuntimeReleaseTypesObject(
	ctx context.Context,
	runtimeRelease platform.RuntimeRelease,
) (types.Object, diag.Diagnostics) {
	obj := RuntimeRelease{
		Version:                  types.StringValue(runtimeRelease.Version),
		AirflowVersion:           types.StringValue(runtimeRelease.AirflowVersion),
		ReleaseDate:              types.StringValue(runtimeRelease.ReleaseDate.String()),
		AirflowDatabaseMigration: types.BoolValue(runtimeRelease.AirflowDatabaseMigration),
		StellarDatabaseMigration: types.BoolValue(runtimeRelease.StellarDatabaseMigration),
		Channel:                  types.StringValue(runtimeRelease.Channel),
	}

	return types.ObjectValueFrom(ctx, schemas.RuntimeReleaseAttributeTypes(), obj)
}

func MachineSpecTypesObject(
	ctx context.Context,
	machineSpec platform.MachineSpec,
) (types.Object, diag.Diagnostics) {
	obj := MachineSpec{
		Cpu:              types.StringValue(machineSpec.Cpu),
		Memory:           types.StringValue(machineSpec.Memory),
		EphemeralStorage: types.StringPointerValue(machineSpec.EphemeralStorage),
	}
	if machineSpec.Concurrency != nil {
		obj.Concurrency = types.StringValue(strconv.FormatFloat(float64(*machineSpec.Concurrency), 'f', -1, 32))
	}

	return types.ObjectValueFrom(ctx, schemas.MachineSpecAttributeTypes(), obj)
}

func SchedulerMachineTypesObject(
	ctx context.Context,
	schedulerMachine platform.SchedulerMachine,
) (types.Object, diag.Diagnostics) {
	spec, diags := MachineSpecTypesObject(ctx, schedulerMachine.Spec)
	if diags.HasError() {
		return types.ObjectNull(schemas.SchedulerMachineAttributeTypes()), diags
	}
	obj := SchedulerMachine{
		Name: types.StringValue(string(schedulerMachine.Name)),
		Spec: spec,
	}

	return types.ObjectValueFrom(ctx, schemas.SchedulerMachineAttributeTypes(), obj)
}

func WorkerMachineTypesObject(
	ctx context.Context,
	workerMachine platform.WorkerMachine,
) (types.Object, diag.Diagnostics) {
	spec, diags := MachineSpecTypesObject(ctx, workerMachine.Spec)
	if diags.HasError() {
		return types.ObjectNull(schemas.MachineSpecAttributeTypes()), diags
	}
	concurrency, diags := RangeTypesObject(ctx, workerMachine.Concurrency)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceRangeAttributeTypes()), diags
	}
	obj := WorkerMachine{
		Name:        types.StringValue(string(workerMachine.Name)),
		Spec:        spec,
		Concurrency: concurrency,
	}

	return types.ObjectValueFrom(ctx, schemas.WorkerMachineAttributeTypes(), obj)
}

func WorkerQueueOptionsTypesObject(
	ctx context.Context,
	workerQueueOptions platform.WorkerQueueOptions,
) (types.Object, diag.Diagnostics) {
	minWorkers, diags := RangeTypesObject(ctx, workerQueueOptions.MinWorkers)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceRangeAttributeTypes()), diags
	}
	maxWorkers, diags := RangeTypesObject(ctx, workerQueueOptions.MaxWorkers)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceRangeAttributeTypes()), diags
	}
	workerConcurrency, diags := RangeTypesObject(ctx, workerQueueOptions.WorkerConcurrency)
	if diags.HasError() {
		return types.ObjectNull(schemas.ResourceRangeAttributeTypes()), diags
	}
	obj := WorkerQueueOptions{
		MinWorkers:        minWorkers,
		MaxWorkers:        maxWorkers,
		WorkerConcurrency: workerConcurrency,
	}

	return types.ObjectValueFrom(ctx, schemas.WorkerQueueOptionsAttributeTypes(), obj)
}

func WorkloadIdentityOptionTypesObject(
	ctx context.Context,
	workloadIdentityOption platform.WorkloadIdentityOption,
) (types.Object, diag.Diagnostics) {
	obj := WorkloadIdentityOption{
		Role:  types.StringValue(workloadIdentityOption.Role),
		Label: types.StringValue(workloadIdentityOption.Label),
	}

	return types.ObjectValueFrom(ctx, schemas.WorkloadIdentityOptionsAttributeTypes(), obj)
}
