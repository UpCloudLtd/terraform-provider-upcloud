package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func serviceDescription(dbType string) string {
	return fmt.Sprintf("This resource represents %s managed database. See UpCloud [Managed Databases](https://upcloud.com/products/managed-databases) product page for more details about the service.", dbType)
}

func setDatabaseValues(ctx context.Context, data *databaseCommonModel, db *upcloud.ManagedDatabase) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	// Title is required, so it should only be empty during import
	isImport := data.Title.ValueString() == ""

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
			Type:   types.StringValue(n.Type),
			Family: types.StringValue(n.Family),
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

	if !data.Properties.IsNull() || isImport {
		respDiagnostics.Append(setDatabaseProperties(ctx, data, db, isImport)...)
	} else {
		data.Properties = types.ListNull(data.Properties.ElementType(ctx))
	}

	return respDiagnostics
}

func ignorePropChange(v any, plan tftypes.Value, key string, prop upcloud.ManagedDatabaseServiceProperty) (any, error) {
	switch key {
	case "ip_filter":
		// API adds /32 postfix to single IP addresses, ignore it when setting value
		p, err := properties.ValueToNative(plan, prop)
		if err != nil {
			return nil, err
		}

		// We already check for null before calling this function, so nil here means unknown value.
		if p == nil {
			return v, nil
		}

		ps, pOk := p.([]any)
		vs, vOk := v.([]any)

		notEqualErr := fmt.Errorf("planned and actual IP filter values do not match: planned %#v, got %#v", p, v)
		if !pOk || !vOk || len(ps) != len(vs) {
			return nil, notEqualErr
		}

		for i := range ps {
			pstr, pOk := ps[i].(string)
			vstr, vOk := vs[i].(string)
			if !pOk || !vOk {
				return nil, notEqualErr
			}

			if pstr != vstr && vstr != pstr+"/32" {
				return nil, notEqualErr
			}
		}
		return p, nil
	default:
		// By default, pass through the current value
		return v, nil
	}
}

func setDatabaseProperties(ctx context.Context, data *databaseCommonModel, db *upcloud.ManagedDatabase, isImport bool) diag.Diagnostics {
	var diags diag.Diagnostics

	propsInfo := properties.GetProperties(db.Type)
	propsData := make(map[string]attr.Value)

	prevProps, err := properties.ListToValueMap(ctx, data.Properties)
	if err != nil {
		diags.AddError(
			"Unable to parse managed database properties from plan",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	for typedKey, value := range db.Properties {
		key := string(typedKey)

		// Skip properties that are not defined in the propsInfo
		prop, ok := propsInfo[key]
		if !ok {
			continue
		}

		// Create-only properties are handled by plan-modifiers
		if prop.CreateOnly {
			continue
		}

		// Skip properties that are null in the plan
		if !isImport && prevProps[properties.SchemaKey(key)].IsNull() {
			continue
		}

		// Ignore known differences between API and plan values
		processedValue, err := ignorePropChange(value, prevProps[properties.SchemaKey(key)], key, prop)
		if err != nil {
			diags.AddError(
				"Unable to process managed database property value",
				utils.ErrorDiagnosticDetail(err),
			)
			return diags
		}

		v, d := properties.NativeToValue(ctx, processedValue, prop)
		diags.Append(d...)

		propsData[properties.SchemaKey(key)], d = properties.ObjectValueAsList(v, prop)
		diags.Append(d...)
	}

	// Clean up removed properties that are not present in the propsInfo
	for key := range propsData {
		if _, ok := propsInfo[key]; !ok {
			delete(propsData, properties.SchemaKey(key))
		}
	}

	// Add null value for properties missing from API response and configuration
	for key, prop := range propsInfo {
		schemaKey := properties.SchemaKey(key)

		// Use value from plan for create-only properties
		if prop.CreateOnly {
			v, d := properties.ValueToAttrValue(ctx, prevProps[schemaKey], prop)
			diags.Append(d...)

			propsData[schemaKey] = v
			continue
		}

		if _, ok := propsData[schemaKey]; !ok {
			nullValue, d := properties.NativeToValue(ctx, nil, propsInfo[key])
			diags.Append(d...)

			propsData[schemaKey], d = properties.ObjectValueAsList(nullValue, propsInfo[key])
			diags.Append(d...)
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

	if err = waitServiceNameToPropagate(ctx, db.ServiceURIParams.Host); err != nil {
		diags.AddWarning(
			"Database DNS name not yet available",
			utils.ErrorDiagnosticDetail(err),
		)
	}

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

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		req.Networks, d = networksFromPlan(ctx, data)
		respDiagnostics.Append(d...)
	}

	if !data.Properties.IsNull() && !data.Properties.IsUnknown() {
		req.Properties, d = buildManagedDatabasePropertiesRequestFromPlan(ctx, data, true)
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

func buildManagedDatabasePropertiesRequestFromPlan(ctx context.Context, data *databaseCommonModel, isCreate bool) (map[upcloud.ManagedDatabasePropertyKey]interface{}, diag.Diagnostics) {
	var respDiagnostics diag.Diagnostics

	dbType := upcloud.ManagedDatabaseServiceType(data.Type.ValueString())
	propsInfo := properties.GetProperties(dbType)

	props, err := properties.PlanToManagedDatabaseProperties(ctx, data.Properties, propsInfo, isCreate)
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

func updateDatabase(ctx context.Context, state, plan *databaseCommonModel, client *service.Service) (*upcloud.ManagedDatabase, string, diag.Diagnostics) {
	var respDiagnostics diag.Diagnostics

	var req request.ModifyManagedDatabaseRequest
	hasChanges := false
	newVersion := ""

	if !state.Plan.Equal(plan.Plan) {
		req.Plan = plan.Plan.ValueString()
		hasChanges = true
	}

	if !state.Title.Equal(plan.Title) {
		req.Title = plan.Title.ValueString()
		hasChanges = true
	}

	if !state.Zone.Equal(plan.Zone) {
		req.Zone = plan.Zone.ValueString()
		hasChanges = true
	}

	if !state.MaintenanceWindowDow.Equal(plan.MaintenanceWindowDow) || !state.MaintenanceWindowTime.Equal(plan.MaintenanceWindowTime) {
		req.Maintenance = request.ManagedDatabaseMaintenanceTimeRequest{
			DayOfWeek: plan.MaintenanceWindowDow.ValueString(),
			Time:      plan.MaintenanceWindowTime.ValueString(),
		}
	}

	if !state.AdditionalDiskSpaceGiB.Equal(plan.AdditionalDiskSpaceGiB) && !plan.AdditionalDiskSpaceGiB.IsNull() && !plan.AdditionalDiskSpaceGiB.IsUnknown() {
		additionalDiskSpaceGiB := int(plan.AdditionalDiskSpaceGiB.ValueInt64())
		req.AdditionalDiskSpaceGiB = &additionalDiskSpaceGiB
		hasChanges = true
	}

	if !state.Labels.Equal(plan.Labels) {
		if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
			var labels map[string]string
			respDiagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
			labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))
			req.Labels = &labelsSlice
			hasChanges = true
		}
	}

	if !state.TerminationProtection.Equal(plan.TerminationProtection) && !plan.TerminationProtection.IsNull() && !plan.TerminationProtection.IsUnknown() {
		terminationProtection := plan.TerminationProtection.ValueBool()
		req.TerminationProtection = &terminationProtection
		hasChanges = true
	}

	if !state.Properties.Equal(plan.Properties) && !plan.Properties.IsNull() {
		props, d := buildManagedDatabasePropertiesRequestFromPlan(ctx, plan, false)
		respDiagnostics.Append(d...)
		stateProps, d := buildManagedDatabasePropertiesRequestFromPlan(ctx, state, false)
		respDiagnostics.Append(d...)

		// Check if version property has changed
		stateVersion := anyToString(stateProps["version"])
		planVersion := anyToString(props["version"])
		if stateVersion != planVersion {
			newVersion = planVersion
		}

		// Always delete version if it exists; versions are updated via separate endpoint
		delete(props, "version")

		req.Properties = props
		hasChanges = true
	}

	if !state.Network.Equal(plan.Network) && !plan.Network.IsNull() {
		networks, d := networksFromPlan(ctx, plan)
		respDiagnostics.Append(d...)

		req.Networks = &networks
		hasChanges = true
	}

	if hasChanges {
		req.UUID = state.ID.ValueString()
		_, err := client.ModifyManagedDatabase(ctx, &req)
		if err != nil {
			respDiagnostics.AddError(
				"Unable to modify managed database",
				utils.ErrorDiagnosticDetail(err),
			)
			return nil, newVersion, respDiagnostics
		}
	}

	if !state.Powered.Equal(plan.Powered) {
		// Wait for non-pending state before updating powered value.
		diags := waitForNonPendingState(ctx, client, plan.ID.ValueString())
		respDiagnostics.Append(diags...)

		respDiagnostics.Append(updatePowered(ctx, plan, client)...)
		if respDiagnostics.HasError() {
			return nil, newVersion, respDiagnostics
		}
	}

	// Wait until database is in running (or stopped) state
	db, diags := waitForPoweredState(ctx, client, plan.ID.ValueString(), plan.Powered.ValueBool())
	respDiagnostics.Append(diags...)

	return db, newVersion, respDiagnostics
}

func updatePowered(ctx context.Context, data *databaseCommonModel, client *service.Service) (diags diag.Diagnostics) {
	var err error
	if data.Powered.ValueBool() {
		_, err = client.StartManagedDatabase(ctx, &request.StartManagedDatabaseRequest{UUID: data.ID.ValueString()})
	} else {
		_, err = client.ShutdownManagedDatabase(ctx, &request.ShutdownManagedDatabaseRequest{UUID: data.ID.ValueString()})
	}
	if err != nil {
		diags.AddError(
			"Unable to modify managed database powered state",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	return diags
}

func anyToString(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func updateVersion(ctx context.Context, uuid, newVersion string, powered bool, client *service.Service) (diags diag.Diagnostics) {
	// Cannot proceed with upgrade if powered off
	if !powered {
		diags.AddError(
			"Unable to upgrade managed database version",
			fmt.Sprintf("Cannot upgrade version for Managed Database %s when it is powered off", uuid),
		)
		return diags
	}

	// Wait until database is in running state before attempting to upgrade version.
	_, err := client.WaitForManagedDatabaseState(ctx, &request.WaitForManagedDatabaseStateRequest{
		UUID:         uuid,
		DesiredState: upcloud.ManagedDatabaseStateRunning,
	})
	if err != nil {
		diags.AddError(
			"Error while waiting for database to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	_, err = client.UpgradeManagedDatabaseVersion(ctx, &request.UpgradeManagedDatabaseVersionRequest{
		UUID:          uuid,
		TargetVersion: newVersion,
	})
	if err != nil {
		diags.AddError(
			"Unable to upgrade managed database version",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	return diags
}

func getDatabaseDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	db, err := svc.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: id[0]})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"resource": "database", "name": db.Name, "state": db.State}, nil
}

func waitForPoweredState(ctx context.Context, svc *service.Service, id string, powered bool) (*upcloud.ManagedDatabase, diag.Diagnostics) {
	var diags diag.Diagnostics

	expectedState := upcloud.ManagedDatabaseStateRunning
	if !powered {
		expectedState = upcloud.ManagedDatabaseStateStopped
	}

	db, err := svc.WaitForManagedDatabaseState(ctx, &request.WaitForManagedDatabaseStateRequest{UUID: id, DesiredState: expectedState})
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error while waiting for database to be in %s state", expectedState),
			utils.ErrorDiagnosticDetail(err),
		)
	}
	return db, diags
}

func waitForDatabaseToBeDeleted(ctx context.Context, svc *service.Service, id string) (diags diag.Diagnostics) {
	err := utils.WaitForResourceToBeDeleted(ctx, svc, getDatabaseDeleted, id)
	if err != nil {
		diags.AddError(
			"Error while waiting for database to be deleted",
			utils.ErrorDiagnosticDetail(err),
		)
	}
	return
}

func waitServiceNameToPropagate(ctx context.Context, name string) (err error) {
	const maxRetries int = 12
	var ips []net.IPAddr
	for i := 0; i <= maxRetries; i++ {
		if ips, err = net.DefaultResolver.LookupIPAddr(ctx, name); err != nil {
			switch e := err.(type) {
			case *net.DNSError:
				if !e.IsNotFound && !e.IsTemporary {
					return err
				}
			default:
				return err
			}
		}

		if len(ips) > 0 {
			return nil
		}

		time.Sleep(10 * time.Second)
	}
	return errors.New("max retries reached while waiting for service name to propagate")
}

func waitForNonPendingState(ctx context.Context, svc *service.Service, uuid string) (diags diag.Diagnostics) {
	for {
		select {
		case <-ctx.Done():
			diags.AddError(
				"Context cancelled",
				utils.ErrorDiagnosticDetail(ctx.Err()),
			)
			return diags
		default:
			db, err := svc.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{
				UUID: uuid,
			})
			if err != nil {
				diags.AddError(
					"Unable to get database details",
					utils.ErrorDiagnosticDetail(err),
				)
				return diags
			}
			if db.State != upcloud.ManagedDatabaseStatePending {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
}
