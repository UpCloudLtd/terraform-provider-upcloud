package database

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func serviceDescription(dbType string) string {
	return fmt.Sprintf("This resource represents %s managed database. See UpCloud [Managed Databases](https://upcloud.com/products/managed-databases) product page for more details about the service.", dbType)
}

func setDatabaseValues(ctx context.Context, data *databaseCommonModel, db *upcloud.ManagedDatabase) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	data.ID = types.StringValue(db.UUID)
	data.Name = types.StringValue(db.Name)
	data.MaintenanceWindowDow = types.StringValue(db.Maintenance.DayOfWeek)
	data.MaintenanceWindowTime = types.StringValue(db.Maintenance.Time)
	data.AdditionalDiskSpaceGiB = types.Int64Value(int64(db.AdditionalDiskSpaceGiB))
	data.Plan = types.StringValue(db.Plan)
	data.Powered = types.BoolValue(db.State == upcloud.ManagedDatabaseStateRunning)
	data.ServiceURI = types.StringValue(db.ServiceURI)
	data.ServiceHost = types.StringValue(db.ServiceURIParams.Host)
	data.ServicePort = types.StringValue(db.ServiceURIParams.Port)
	data.ServiceUsername = types.StringValue(db.ServiceURIParams.User)
	data.ServicePassword = types.StringValue(db.ServiceURIParams.Password)
	data.State = types.StringValue(string(db.State))
	data.TerminationProtection = types.BoolValue(db.TerminationProtection)
	data.Title = types.StringValue(db.Title)
	data.Type = types.StringValue(string(db.Type))
	data.Zone = types.StringValue(db.Zone)
	data.PrimaryDatabase = types.StringValue(db.ServiceURIParams.DatabaseName)

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(db.Labels))
	respDiagnostics.Append(diags...)

	var components []databaseComponentModel
	for _, c := range db.Components {
		components = append(components, databaseComponentModel{
			Component: types.StringValue(c.Component),
			Host:      types.StringValue(c.Host),
			Port:      types.Int64Value(int64(c.Port)),
			Route:     types.StringValue(string(c.Route)),
			Usage:     types.StringValue(string(c.Usage)),
		})
	}
	data.Components, diags = types.ListValueFrom(ctx, data.Components.ElementType(ctx), components)
	respDiagnostics.Append(diags...)

	var networks []databaseNetworkModel
	for _, n := range db.Networks {
		networks = append(networks, databaseNetworkModel{
			Name:   types.StringValue(n.Name),
			Type:   types.StringValue(string(n.Type)),
			Family: types.StringValue(string(n.Family)),
			UUID:   types.StringPointerValue(n.UUID),
		})
	}
	data.Network, diags = types.SetValueFrom(ctx, data.Network.ElementType(ctx), networks)
	respDiagnostics.Append(diags...)

	var nodeStates []databaseNodeStateModel
	for _, ns := range db.NodeStates {
		nodeStates = append(nodeStates, databaseNodeStateModel{
			Name:  types.StringValue(ns.Name),
			Role:  types.StringValue(string(ns.Role)),
			State: types.StringValue(ns.State),
		})
	}
	data.NodeStates, diags = types.ListValueFrom(ctx, data.NodeStates.ElementType(ctx), nodeStates)
	respDiagnostics.Append(diags...)

	respDiagnostics.Append(setDatabaseProperties(ctx, data, db)...)

	return diags
}

func setDatabaseProperties(ctx context.Context, data *databaseCommonModel, db *upcloud.ManagedDatabase) diag.Diagnostics {
	var diags diag.Diagnostics

	propsInfo := properties.GetProperties(db.Type)
	propsData := make(map[string]attr.Value)

	for typedKey, value := range db.Properties {
		key := string(typedKey)

		// Create-only properties are handled by plan-modifiers
		if propsInfo[key].CreateOnly {
			continue
		}

		if properties.GetType(propsInfo[key]) == "object" {
			// convert API objects into list of objects
			if m, ok := value.(map[string]interface{}); ok {
				// Convert API keys to schema keys
				sMap := make(map[string]interface{})
				for k, v := range m {
					sMap[properties.SchemaKey(k)] = v
				}
				o, d := properties.NativeToValue(ctx, sMap, propsInfo[key])
				diags.Append(d...)

				propsData[properties.SchemaKey(key)], d = types.ListValueFrom(ctx, properties.PropToAttributeType(propsInfo[key]), []attr.Value{o})
			}
		} else {
			v, d := properties.NativeToValue(ctx, value, propsInfo[key])
			diags.Append(d...)
			propsData[properties.SchemaKey(key)] = v
		}
	}

	// Clean up removed properties that are not present in the propsInfo
	for key := range propsData {
		if _, ok := propsInfo[key]; !ok {
			delete(propsData, properties.SchemaKey(key))
		}
	}

	// Add null value for properties missing from API response and configuration
	for key := range propsInfo {
		schemaKey := properties.SchemaKey(key)
		if _, ok := propsData[schemaKey]; !ok {
			nullValue, d := properties.NativeToValue(ctx, nil, propsInfo[key])
			diags.Append(d...)
			propsData[schemaKey] = nullValue
		}
	}

	props, d := types.ObjectValue(properties.PropsToAttributeTypes(propsInfo), propsData)
	diags.Append(d...)

	data.Properties, d = types.ListValue(data.Properties.ElementType(ctx), []attr.Value{props})
	diags.Append(d...)

	return diags
}

func createDatabase(ctx context.Context, data *databaseCommonModel, client *service.Service) (*upcloud.ManagedDatabase, diag.Diagnostics) {
	var diags diag.Diagnostics

	req, d := buildManagedDatabaseRequestFromPlan(ctx, data)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	db, err := client.CreateManagedDatabase(ctx, &req)
	if err != nil {
		diags.AddError(
			"Unable to create database",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	data.ID = types.StringValue(db.UUID)

	if db, err = client.WaitForManagedDatabaseState(ctx, &request.WaitForManagedDatabaseStateRequest{UUID: db.UUID, DesiredState: upcloud.ManagedDatabaseStateRunning}); err != nil {
		diags.AddError(
			"Error while waiting for database to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	diags.Append(setDatabaseValues(ctx, data, db)...)
	return db, diags
}

func buildManagedDatabaseRequestFromPlan(ctx context.Context, data *databaseCommonModel) (request.CreateManagedDatabaseRequest, diag.Diagnostics) {
	var d, respDiagnostics diag.Diagnostics

	var terminationProtection *bool
	if !data.TerminationProtection.IsNull() {
		tp := data.TerminationProtection.ValueBool()
		terminationProtection = &tp
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		respDiagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	req := request.CreateManagedDatabaseRequest{
		AdditionalDiskSpaceGiB: int(data.AdditionalDiskSpaceGiB.ValueInt64()),
		HostNamePrefix:         data.Name.ValueString(),
		Plan:                   data.Plan.ValueString(),
		Title:                  data.Title.ValueString(),
		TerminationProtection:  terminationProtection,
		Labels:                 utils.LabelsMapToSlice(labels),
		Type:                   upcloud.ManagedDatabaseServiceType(data.Type.ValueString()),
		Zone:                   data.Zone.ValueString(),
	}

	if data.Network.IsNull() || data.Network.IsUnknown() {
		req.Networks, d = networksFromPlan(ctx, data)
		respDiagnostics.Append(d...)
	}

	if data.Properties.IsNull() || data.Properties.IsUnknown() {
		req.Properties, d = buildManagedDatabasePropertiesRequestFromPlan(ctx, data)
		respDiagnostics.Append(d...)
	}

	if data.MaintenanceWindowDow.ValueString() != "" && data.MaintenanceWindowTime.ValueString() != "" {
		req.Maintenance = request.ManagedDatabaseMaintenanceTimeRequest{
			DayOfWeek: data.MaintenanceWindowDow.ValueString(),
			Time:      data.MaintenanceWindowTime.ValueString(),
		}
	}

	return req, respDiagnostics
}

func buildManagedDatabasePropertiesRequestFromPlan(ctx context.Context, data *databaseCommonModel) (map[upcloud.ManagedDatabasePropertyKey]interface{}, diag.Diagnostics) {
	var respDiagnostics diag.Diagnostics

	dbType := upcloud.ManagedDatabaseServiceType(data.Type.ValueString())
	propsInfo := properties.GetProperties(dbType)

	props, err := properties.PlanToManagedDatabaseProperties(ctx, data.Properties, propsInfo)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to build managed database properties from plan",
			utils.ErrorDiagnosticDetail(err),
		)
	}
	return props, respDiagnostics
}

func readDatabase(ctx context.Context, data *databaseCommonModel, client *service.Service, removeFromState func(context.Context)) (*upcloud.ManagedDatabase, diag.Diagnostics) {
	var diags diag.Diagnostics

	if data.ID.ValueString() == "" {
		removeFromState(ctx)
		return nil, diags
	}

	db, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			removeFromState(ctx)
		} else {
			diags.AddError(
				"Unable to read managed database details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return nil, diags
	}

	diags.Append(setDatabaseValues(ctx, data, db)...)
	return db, diags
}

func waitForDatabaseToBeDeletedDiags(ctx context.Context, svc *service.Service, id string) (diags diag.Diagnostics) {
	err := utils.WaitForResourceToBeDeleted(ctx, svc, getDatabaseDeleted, id)
	if err != nil {
		diags.AddError(
			"Error while waiting for database to be deleted",
			utils.ErrorDiagnosticDetail(err),
		)
	}
	return
}
