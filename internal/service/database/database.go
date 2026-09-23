package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	databaseStateRunning = "running"
	databaseStateStopped = "stopped"
	databaseStatePending = "pending"
	databaseStateError   = "error"

	databasePollInterval        = 5 * time.Second
	maxDatabasePollServerErrors = 3
)

func serviceDescription(dbType string) string {
	return fmt.Sprintf("This resource represents %s managed database. See UpCloud [Managed Databases](https://upcloud.com/products/managed-databases) product page for more details about the service.", dbType)
}

func serviceURIParamString(params *map[string]interface{}, key string) string {
	if params == nil {
		return ""
	}
	v, ok := (*params)[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.Itoa(int(t))
	default:
		return fmt.Sprintf("%v", t)
	}
}

func setDatabaseValues(ctx context.Context, data *databaseCommonModel, db *v9.DatabaseServiceInformationResponse, isImport bool) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	if db.Uuid != nil {
		data.ID = types.StringValue(db.Uuid.String())
	}
	if db.Name != nil {
		data.Name = types.StringPointerValue(db.Name)
	}
	if db.Maintenance != nil {
		data.MaintenanceWindowDow = types.StringPointerValue(db.Maintenance.Dow)
		data.MaintenanceWindowTime = types.StringPointerValue(db.Maintenance.Time)
	} else if isImport {
		data.MaintenanceWindowDow = types.StringNull()
		data.MaintenanceWindowTime = types.StringNull()
	}
	if db.AdditionalDiskSpaceGib != nil {
		data.AdditionalDiskSpaceGiB = types.Int64Value(int64(*db.AdditionalDiskSpaceGib))
	} else if isImport {
		data.AdditionalDiskSpaceGiB = types.Int64Null()
	}
	if db.Plan != nil {
		data.Plan = types.StringPointerValue(db.Plan)
	} else if isImport {
		data.Plan = types.StringNull()
	}
	if db.Powered != nil {
		data.Powered = types.BoolPointerValue(db.Powered)
	} else if db.State != nil {
		data.Powered = types.BoolValue(*db.State == databaseStateRunning)
	} else if isImport {
		data.Powered = types.BoolNull()
	}
	data.ServiceURI = types.StringPointerValue(db.ServiceUri)
	data.ServiceHost = types.StringValue(serviceURIParamString(db.ServiceUriParams, "host"))
	data.ServicePort = types.StringValue(serviceURIParamString(db.ServiceUriParams, "port"))
	data.ServiceUsername = types.StringValue(serviceURIParamString(db.ServiceUriParams, "user"))
	data.ServicePassword = types.StringValue(serviceURIParamString(db.ServiceUriParams, "password"))
	data.State = types.StringPointerValue(db.State)
	if db.TerminationProtection != nil {
		data.TerminationProtection = types.BoolPointerValue(db.TerminationProtection)
	} else if isImport {
		data.TerminationProtection = types.BoolNull()
	}
	if db.Title != nil {
		data.Title = types.StringPointerValue(db.Title)
	}
	if db.Type != nil {
		data.Type = types.StringPointerValue(db.Type)
	}
	if db.Zone != nil {
		data.Zone = types.StringPointerValue(db.Zone)
	}
	data.PrimaryDatabase = types.StringValue(serviceURIParamString(db.ServiceUriParams, "dbname"))

	if db.Labels != nil {
		labels := make(map[string]string, len(*db.Labels))
		for _, l := range *db.Labels {
			if l.Key != nil && l.Value != nil {
				labels[*l.Key] = *l.Value
			}
		}
		data.Labels, diags = types.MapValueFrom(ctx, types.StringType, labels)
		respDiagnostics.Append(diags...)
	} else if isImport {
		data.Labels = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	var components []databaseComponentModel
	if db.Components != nil {
		for _, c := range *db.Components {
			component := databaseComponentModel{
				Component: types.StringPointerValue(c.Component),
				Host:      types.StringPointerValue(c.Host),
				Route:     types.StringPointerValue(c.Route),
				Usage:     types.StringPointerValue(c.Usage),
			}
			if c.Port != nil {
				component.Port = types.Int64Value(int64(*c.Port))
			}
			components = append(components, component)
		}
	}
	data.Components, diags = types.ListValueFrom(ctx, data.Components.ElementType(ctx), components)
	respDiagnostics.Append(diags...)

	if db.Networks != nil {
		var networks []databaseNetworkModel
		for _, n := range *db.Networks {
			network := databaseNetworkModel{
				Name: types.StringPointerValue(n.Name),
			}
			if n.Type != nil {
				network.Type = types.StringValue(string(*n.Type))
			}
			if n.Family != nil {
				network.Family = types.StringValue(string(*n.Family))
			}
			if n.Uuid != nil {
				uuidStr := n.Uuid.String()
				network.UUID = types.StringValue(uuidStr)
			}
			networks = append(networks, network)
		}
		data.Network, diags = types.SetValueFrom(ctx, data.Network.ElementType(ctx), networks)
		respDiagnostics.Append(diags...)
	} else if isImport {
		data.Network = types.SetValueMust(data.Network.ElementType(ctx), []attr.Value{})
	}

	var nodeStates []databaseNodeStateModel
	if db.NodeStates != nil {
		for _, ns := range *db.NodeStates {
			nodeStates = append(nodeStates, databaseNodeStateModel{
				Name:  types.StringPointerValue(ns.Name),
				Role:  types.StringPointerValue(ns.Role),
				State: types.StringPointerValue(ns.State),
			})
		}
	}
	data.NodeStates, diags = types.ListValueFrom(ctx, data.NodeStates.ElementType(ctx), nodeStates)
	respDiagnostics.Append(diags...)

	if db.Properties == nil && !isImport {
		return respDiagnostics
	}
	if !data.Properties.IsNull() || isImport {
		var props map[string]interface{}
		if db.Properties != nil {
			props = *db.Properties
		}
		dbType := ""
		if db.Type != nil {
			dbType = *db.Type
		}
		respDiagnostics.Append(setDatabaseProperties(ctx, data, props, dbType, isImport)...)
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

		if !pOk || !vOk || len(ps) != len(vs) {
			// Lengths differ (e.g. extra IPs added outside Terraform): return actual value
			// so Terraform can plan to reconcile the difference.
			return v, nil
		}

		for i := range ps {
			pstr, pOk := ps[i].(string)
			vstr, vOk := vs[i].(string)
			if !pOk || !vOk {
				return v, nil
			}

			if pstr != vstr && vstr != pstr+"/32" {
				// Values differ beyond the /32 suffix normalisation: return actual value.
				return v, nil
			}
		}
		return p, nil
	default:
		// By default, pass through the current value
		return v, nil
	}
}

func setDatabaseProperties(ctx context.Context, data *databaseCommonModel, dbProps map[string]interface{}, dbType string, isImport bool) diag.Diagnostics {
	var diags, d diag.Diagnostics

	propsInfo := properties.GetProperties(upcloud.ManagedDatabaseServiceType(dbType))
	propsData := make(map[string]attr.Value)

	prevProps, err := properties.ListToValueMap(ctx, data.Properties)
	if err != nil {
		diags.AddError(
			"Unable to parse managed database properties from plan",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	for key, value := range dbProps {
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

		propsData[properties.SchemaKey(key)], d = properties.NativeToValue(ctx, processedValue, prop)
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
			propsData[schemaKey], d = properties.NativeToValue(ctx, nil, propsInfo[key])
			diags.Append(d...)
		}
	}

	props, d := types.ObjectValue(properties.PropsToAttributeTypes(propsInfo), propsData)
	diags.Append(d...)

	data.Properties, d = types.ListValue(data.Properties.ElementType(ctx), []attr.Value{props})
	diags.Append(d...)

	return diags
}

func createDatabase(ctx context.Context, data *databaseCommonModel, componentPlan *databasePlanModel, client *v9.ClientWithResponses) (*v9.DatabaseServiceInformationResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	req, d := buildManagedDatabaseRequestFromPlan(ctx, data, componentPlan)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	apiResp, err := client.CreateDatabaseWithResponse(ctx, req)
	if err != nil {
		diags.AddError(
			"Unable to create database",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}
	if apiResp.StatusCode() != http.StatusCreated || apiResp.JSON201 == nil || apiResp.JSON201.Uuid == nil {
		diags.AddError(
			"Unable to create database",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return nil, diags
	}

	uuid := apiResp.JSON201.Uuid.String()
	data.ID = types.StringValue(uuid)
	if componentPlan != nil {
		if componentPlan.PlanCompute.IsNull() || componentPlan.PlanCompute.IsUnknown() {
			clearDatabasePlanComponents(componentPlan)
		}
		setDatabasePlanComponents(componentPlan, apiResp.JSON201.PlanComponents)
	}

	db, d := waitForDatabaseState(ctx, client, uuid, databaseStateRunning)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	diags.Append(setDatabaseValues(ctx, data, db, false)...)

	host := serviceURIParamString(db.ServiceUriParams, "host")
	if err = waitServiceNameToPropagate(ctx, host); err != nil {
		diags.AddWarning(
			"Database DNS name not yet available",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	return db, diags
}

func buildManagedDatabaseRequestFromPlan(ctx context.Context, data *databaseCommonModel, componentPlan *databasePlanModel) (v9.CreateDatabaseJSONRequestBody, diag.Diagnostics) {
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

	req := v9.CreateDatabaseJSONRequestBody{
		HostnamePrefix:        data.Name.ValueString(),
		Title:                 data.Title.ValueString(),
		TerminationProtection: terminationProtection,
		Type:                  v9.DatabaseServiceType(data.Type.ValueString()),
		Zone:                  data.Zone.ValueString(),
	}

	if labels != nil {
		dbLabels := make([]v9.DatabaseLabelCreate, 0, len(labels))
		for k, v := range labels {
			dbLabels = append(dbLabels, v9.DatabaseLabelCreate{Key: k, Value: v})
		}
		req.Labels = &dbLabels
	}

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		req.Networks, d = networksV9FromPlan(ctx, data)
		respDiagnostics.Append(d...)
	}

	if !data.Properties.IsNull() && !data.Properties.IsUnknown() {
		props, d := buildManagedDatabasePropertiesRequestFromPlan(ctx, data, true)
		respDiagnostics.Append(d...)

		propsBody, err := json.Marshal(props)
		if err == nil {
			var dbProps v9.DatabaseServiceCreateOpenAPI_Properties
			err = json.Unmarshal(propsBody, &dbProps)
			req.Properties = &dbProps
		}
		if err != nil {
			respDiagnostics.AddError("Unable to build managed database request", utils.ErrorDiagnosticDetail(err))
			return v9.CreateDatabaseJSONRequestBody{}, respDiagnostics
		}
	}

	if data.MaintenanceWindowDow.ValueString() != "" && data.MaintenanceWindowTime.ValueString() != "" {
		req.Maintenance = &struct {
			Dow  v9.DatabaseMaintenanceDow  `json:"dow"`
			Time v9.DatabaseMaintenanceTime `json:"time"`
		}{
			Dow:  v9.DatabaseMaintenanceDow(data.MaintenanceWindowDow.ValueString()),
			Time: data.MaintenanceWindowTime.ValueString(),
		}
	}

	if componentPlan != nil && !componentPlan.PlanCompute.IsNull() && !componentPlan.PlanCompute.IsUnknown() {
		req.PlanCompute = componentPlan.PlanCompute.ValueStringPointer()
		nodeCount := int(componentPlan.PlanNodeCount.ValueInt64())
		req.PlanNodeCount = &nodeCount
		storageGiB := int(componentPlan.PlanStorageGiB.ValueInt64())
		req.PlanStorageGib = &storageGiB
		backups := v9.DatabaseServiceCreateOpenAPIPlanBackups(componentPlan.PlanBackups.ValueString())
		req.PlanBackups = &backups
	} else {
		plan := data.Plan.ValueString()
		req.Plan = &plan
		additionalDiskSpaceGiB := int(data.AdditionalDiskSpaceGiB.ValueInt64())
		req.AdditionalDiskSpaceGib = &additionalDiskSpaceGiB
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

func readDatabase(ctx context.Context, data *databaseCommonModel, componentPlan *databasePlanModel, client *v9.ClientWithResponses, removeFromState func(context.Context)) (*v9.DatabaseServiceInformationResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	if data.ID.ValueString() == "" {
		removeFromState(ctx)
		return nil, diags
	}

	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		diags.AddError(
			"Unable to read managed database details",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	apiResp, err := client.GetDatabaseWithResponse(ctx, serviceUUID)
	if err != nil {
		diags.AddError(
			"Unable to read managed database details",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		removeFromState(ctx)
		return nil, diags
	}
	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		diags.AddError(
			"Unable to read managed database details",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return nil, diags
	}

	db := apiResp.JSON200
	isImport := data.Title.IsNull() || data.Title.ValueString() == ""
	diags.Append(setDatabaseValues(ctx, data, db, isImport)...)

	if componentPlan != nil {
		if isImport {
			clearDatabasePlanComponents(componentPlan)
		}
		setDatabasePlanComponents(componentPlan, db.PlanComponents)
	}

	return db, diags
}

func clearDatabasePlanComponents(data *databasePlanModel) {
	data.PlanBackups = types.StringNull()
	data.PlanCompute = types.StringNull()
	data.PlanNodeCount = types.Int64Null()
	data.PlanStorageGiB = types.Int64Null()
}

func setDatabasePlanComponents(data *databasePlanModel, components *v9.DatabasePlanComponentsResponse) {
	if components == nil {
		return
	}

	if components.Compute != nil {
		if components.Compute.Name != nil {
			data.PlanCompute = types.StringValue(*components.Compute.Name)
		}
		if components.Compute.NodeCount != nil {
			data.PlanNodeCount = types.Int64Value(int64(*components.Compute.NodeCount))
		}
	}
	if components.Storage != nil && components.Storage.TotalGib != nil {
		data.PlanStorageGiB = types.Int64Value(int64(*components.Storage.TotalGib))
	}
	if components.Backups != nil && components.Backups.Name != nil {
		data.PlanBackups = types.StringValue(string(*components.Backups.Name))
	}
}

func resolveUnknownDatabasePlanComponents(plan, state *databasePlanModel) {
	if plan.PlanCompute.IsUnknown() {
		plan.PlanCompute = state.PlanCompute
	}
	if plan.PlanNodeCount.IsUnknown() {
		plan.PlanNodeCount = state.PlanNodeCount
	}
	if plan.PlanStorageGiB.IsUnknown() {
		plan.PlanStorageGiB = state.PlanStorageGiB
	}
	if plan.PlanBackups.IsUnknown() {
		plan.PlanBackups = state.PlanBackups
	}
}

func databaseComponentPlanChanged(state, plan *databasePlanModel) bool {
	return !state.PlanCompute.Equal(plan.PlanCompute) ||
		!state.PlanNodeCount.Equal(plan.PlanNodeCount) ||
		!state.PlanStorageGiB.Equal(plan.PlanStorageGiB) ||
		!state.PlanBackups.Equal(plan.PlanBackups)
}

func buildManagedDatabaseModifyRequestFromPlan(ctx context.Context, state, plan, config *databaseCommonModel, stateComponentPlan, plannedComponentPlan, configComponentPlan *databasePlanModel) (v9.ModifyDatabaseJSONRequestBody, bool, string, diag.Diagnostics) {
	var respDiagnostics diag.Diagnostics

	var req v9.ModifyDatabaseJSONRequestBody
	hasChanges := false
	newVersion := ""

	if !config.Plan.IsNull() && !config.Plan.IsUnknown() && !state.Plan.Equal(plan.Plan) {
		req.Plan = plan.Plan.ValueStringPointer()
		hasChanges = true
	}

	componentPlanConfigured := configComponentPlan != nil && !configComponentPlan.PlanCompute.IsNull() && !configComponentPlan.PlanCompute.IsUnknown()
	componentPlanChanged := stateComponentPlan != nil && plannedComponentPlan != nil && databaseComponentPlanChanged(stateComponentPlan, plannedComponentPlan)
	if componentPlanConfigured && componentPlanChanged {
		nodeCount := int(plannedComponentPlan.PlanNodeCount.ValueInt64())
		storageGiB := int(plannedComponentPlan.PlanStorageGiB.ValueInt64())
		backups := v9.DatabaseServiceModifyOpenAPIPlanBackups(plannedComponentPlan.PlanBackups.ValueString())
		req.PlanCompute = plannedComponentPlan.PlanCompute.ValueStringPointer()
		req.PlanNodeCount = &nodeCount
		req.PlanStorageGib = &storageGiB
		req.PlanBackups = &backups
		hasChanges = true
	}

	if !state.Title.Equal(plan.Title) {
		req.Title = plan.Title.ValueStringPointer()
		hasChanges = true
	}

	if !state.Zone.Equal(plan.Zone) {
		req.Zone = plan.Zone.ValueStringPointer()
		hasChanges = true
	}

	maintenanceConfigured := !config.MaintenanceWindowDow.IsNull() &&
		!config.MaintenanceWindowDow.IsUnknown() &&
		!config.MaintenanceWindowTime.IsNull() &&
		!config.MaintenanceWindowTime.IsUnknown()
	maintenanceChanged := !state.MaintenanceWindowDow.Equal(plan.MaintenanceWindowDow) ||
		!state.MaintenanceWindowTime.Equal(plan.MaintenanceWindowTime)
	if maintenanceConfigured && maintenanceChanged {
		req.Maintenance = &struct {
			Dow  v9.DatabaseMaintenanceDow  `json:"dow"`
			Time v9.DatabaseMaintenanceTime `json:"time"`
		}{
			Dow:  v9.DatabaseMaintenanceDow(plan.MaintenanceWindowDow.ValueString()),
			Time: plan.MaintenanceWindowTime.ValueString(),
		}
		hasChanges = true
	}

	if !state.AdditionalDiskSpaceGiB.Equal(plan.AdditionalDiskSpaceGiB) && !plan.AdditionalDiskSpaceGiB.IsNull() && !plan.AdditionalDiskSpaceGiB.IsUnknown() {
		additionalDiskSpaceGiB := int(plan.AdditionalDiskSpaceGiB.ValueInt64())
		req.AdditionalDiskSpaceGib = &additionalDiskSpaceGiB
		hasChanges = true
	}

	if !state.Labels.Equal(plan.Labels) {
		if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
			var labels map[string]string
			respDiagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
			labelsSlice := make([]v9.DatabaseLabelCreate, 0, len(labels))
			for k, v := range labels {
				labelsSlice = append(labelsSlice, v9.DatabaseLabelCreate{Key: k, Value: v})
			}
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

		propsBody, err := json.Marshal(props)
		if err == nil {
			var dbProps v9.DatabaseServiceModifyOpenAPI_Properties
			err = json.Unmarshal(propsBody, &dbProps)
			req.Properties = &dbProps
		}
		if err != nil {
			respDiagnostics.AddError("Unable to modify managed database", utils.ErrorDiagnosticDetail(err))
			return v9.ModifyDatabaseJSONRequestBody{}, false, newVersion, respDiagnostics
		}
		hasChanges = true
	}

	if !state.Network.Equal(plan.Network) && !plan.Network.IsNull() {
		networks, d := networksV9FromPlan(ctx, plan)
		respDiagnostics.Append(d...)

		req.Networks = networks
		hasChanges = true
	}

	return req, hasChanges, newVersion, respDiagnostics
}

func updateDatabase(ctx context.Context, state, plan, config *databaseCommonModel, stateComponentPlan, plannedComponentPlan, configComponentPlan *databasePlanModel, client *v9.ClientWithResponses) (*v9.DatabaseServiceInformationResponse, string, diag.Diagnostics) {
	req, hasChanges, newVersion, respDiagnostics := buildManagedDatabaseModifyRequestFromPlan(
		ctx,
		state,
		plan,
		config,
		stateComponentPlan,
		plannedComponentPlan,
		configComponentPlan,
	)
	if respDiagnostics.HasError() {
		return nil, newVersion, respDiagnostics
	}

	serviceUUID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		respDiagnostics.AddError("Unable to modify managed database", utils.ErrorDiagnosticDetail(err))
		return nil, newVersion, respDiagnostics
	}

	if hasChanges {
		apiResp, err := client.ModifyDatabaseWithResponse(ctx, serviceUUID, req)
		if err != nil {
			respDiagnostics.AddError(
				"Unable to modify managed database",
				utils.ErrorDiagnosticDetail(err),
			)
			return nil, newVersion, respDiagnostics
		}
		if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
			respDiagnostics.AddError(
				"Unable to modify managed database",
				fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
			)
			return nil, newVersion, respDiagnostics
		}
	}

	if !state.Powered.Equal(plan.Powered) {
		// Wait for non-pending state before updating powered value.
		diags := waitForNonPendingState(ctx, client, plan.ID.ValueString())
		respDiagnostics.Append(diags...)
		if respDiagnostics.HasError() {
			return nil, newVersion, respDiagnostics
		}

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

func updatePowered(ctx context.Context, data *databaseCommonModel, client *v9.ClientWithResponses) (diags diag.Diagnostics) {
	serviceUUID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		diags.AddError(
			"Unable to modify managed database powered state",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	powered := data.Powered.ValueBool()
	apiResp, err := client.ModifyDatabaseWithResponse(ctx, serviceUUID, v9.ModifyDatabaseJSONRequestBody{
		Powered: &powered,
	})
	if err != nil {
		diags.AddError(
			"Unable to modify managed database powered state",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}
	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		diags.AddError(
			"Unable to modify managed database powered state",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
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

func databaseTaskError(result *string, resultCodes *[]map[string]string) string {
	details := make([]string, 0)
	if result != nil && *result != "" {
		details = append(details, *result)
	}
	if resultCodes != nil {
		for _, resultCode := range *resultCodes {
			keys := make([]string, 0, len(resultCode))
			for key := range resultCode {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				details = append(details, fmt.Sprintf("%s: %s", key, resultCode[key]))
			}
		}
	}
	if len(details) == 0 {
		return "database upgrade check failed"
	}
	return strings.Join(details, "; ")
}

func getDatabaseTaskByUUID(ctx context.Context, client *v9.ClientWithResponses, serviceUUID, taskUUID uuid.UUID) (*v9.DatabaseServiceTaskDetailsResponse, error) {
	taskURL := fmt.Sprintf(
		"%s1.3/database/%s/tasks/%s",
		client.Server,
		url.PathEscape(serviceUUID.String()),
		url.PathEscape(taskUUID.String()),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, taskURL, nil)
	if err != nil {
		return nil, err
	}
	for _, editor := range client.RequestEditors {
		if err := editor(ctx, req); err != nil {
			return nil, err
		}
	}

	resp, err := client.Client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected API status code %d: %s", resp.StatusCode, string(body))
	}

	var task v9.DatabaseServiceTaskDetailsResponse
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func waitForDatabaseUpgradeCheck(ctx context.Context, client *v9.ClientWithResponses, serviceUUID uuid.UUID, task *v9.DatabaseServiceTaskResponse, interval time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	if task.Success != nil {
		if !*task.Success {
			diags.AddError(
				"Managed database upgrade check failed",
				databaseTaskError(task.Result, task.ResultCodes),
			)
		}
		return diags
	}
	if task.Id == nil {
		diags.AddError(
			"Unable to run managed database upgrade check",
			"API returned an asynchronous upgrade check without a task ID",
		)
		return diags
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			diags.AddError(
				"Context cancelled while waiting for database upgrade check",
				utils.ErrorDiagnosticDetail(ctx.Err()),
			)
			return diags
		case <-ticker.C:
			taskDetails, err := getDatabaseTaskByUUID(ctx, client, serviceUUID, *task.Id)
			if err != nil {
				diags.AddError(
					"Unable to get managed database upgrade check result",
					utils.ErrorDiagnosticDetail(err),
				)
				return diags
			}
			if taskDetails.Success == nil {
				continue
			}
			if !*taskDetails.Success {
				diags.AddError(
					"Managed database upgrade check failed",
					databaseTaskError(taskDetails.Result, taskDetails.ResultCodes),
				)
			}
			return diags
		}
	}
}

func runDatabaseUpgradeCheck(ctx context.Context, client *v9.ClientWithResponses, serviceUUID uuid.UUID, targetVersion string) diag.Diagnostics {
	var diags diag.Diagnostics

	apiResp, err := client.CreateDatabaseTaskWithResponse(ctx, serviceUUID, v9.CreateDatabaseTaskJSONRequestBody{
		Operation: "upgrade_check",
		UpgradeCheck: &struct {
			TargetVersion string `json:"target_version"`
		}{
			TargetVersion: targetVersion,
		},
	})
	if err != nil {
		diags.AddError(
			"Unable to run managed database upgrade check",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}
	if apiResp.StatusCode() != http.StatusCreated || apiResp.JSON201 == nil {
		diags.AddError(
			"Unable to run managed database upgrade check",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return diags
	}

	return waitForDatabaseUpgradeCheck(ctx, client, serviceUUID, apiResp.JSON201, databasePollInterval)
}

func updateVersion(ctx context.Context, uuidStr, newVersion string, powered bool, client *v9.ClientWithResponses) (diags diag.Diagnostics) {
	// Cannot proceed with upgrade if powered off
	if !powered {
		diags.AddError(
			"Unable to upgrade managed database version",
			fmt.Sprintf("Cannot upgrade version for Managed Database %s when it is powered off", uuidStr),
		)
		return diags
	}

	// Wait until database is in running state before attempting to upgrade version.
	_, d := waitForDatabaseState(ctx, client, uuidStr, databaseStateRunning)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	serviceUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		diags.AddError(
			"Unable to upgrade managed database version",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	diags.Append(runDatabaseUpgradeCheck(ctx, client, serviceUUID, newVersion)...)
	if diags.HasError() {
		return diags
	}

	apiResp, err := client.UpgradeDatabaseWithResponse(ctx, serviceUUID, v9.UpgradeDatabaseJSONRequestBody{
		TargetVersion: newVersion,
	})
	if err != nil {
		diags.AddError(
			"Unable to upgrade managed database version",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}
	if apiResp.StatusCode() != http.StatusCreated || apiResp.JSON201 == nil {
		diags.AddError(
			"Unable to upgrade managed database version",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
	}

	return diags
}

func databaseStateErrorDetails(stateErrors *map[string]string) string {
	if stateErrors == nil || len(*stateErrors) == 0 {
		return "managed database entered error state"
	}

	keys := make([]string, 0, len(*stateErrors))
	for key := range *stateErrors {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	details := make([]string, 0, len(keys))
	for _, key := range keys {
		details = append(details, fmt.Sprintf("%s (%s)", (*stateErrors)[key], key))
	}
	return fmt.Sprintf("managed database entered error state: %s", strings.Join(details, ", "))
}

// pollDatabase polls the database until match returns true for its state, the get
// fails with a not found status and matchNotFound is true, or the context is done.
func pollDatabase(ctx context.Context, client *v9.ClientWithResponses, id string, match func(state string) bool, matchNotFound bool) (*v9.DatabaseServiceInformationResponse, diag.Diagnostics) {
	return pollDatabaseWithInterval(ctx, client, id, match, matchNotFound, databasePollInterval)
}

func pollDatabaseWithInterval(ctx context.Context, client *v9.ClientWithResponses, id string, match func(state string) bool, matchNotFound bool, interval time.Duration) (*v9.DatabaseServiceInformationResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	serviceUUID, err := uuid.Parse(id)
	if err != nil {
		diags.AddError(
			"Unable to parse database UUID",
			utils.ErrorDiagnosticDetail(err),
		)
		return nil, diags
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	serverErrors := 0
	for {
		apiResp, err := client.GetDatabaseWithResponse(ctx, serviceUUID)
		if err != nil {
			if ctx.Err() != nil {
				diags.AddError(
					"Context cancelled while waiting for database state",
					utils.ErrorDiagnosticDetail(ctx.Err()),
				)
				return nil, diags
			}
			diags.AddError(
				"Unable to get database details while waiting for state",
				utils.ErrorDiagnosticDetail(err),
			)
			return nil, diags
		}

		switch {
		case apiResp.StatusCode() == http.StatusNotFound:
			if matchNotFound {
				return nil, diags
			}
			diags.AddError(
				"Unable to get database details while waiting for state",
				fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
			)
			return nil, diags
		case apiResp.StatusCode() >= http.StatusInternalServerError:
			serverErrors++
			if serverErrors >= maxDatabasePollServerErrors {
				diags.AddError(
					"Unable to get database details while waiting for state",
					fmt.Sprintf("API returned status code %d after %d attempts: %s", apiResp.StatusCode(), serverErrors, string(apiResp.Body)),
				)
				return nil, diags
			}
		case apiResp.StatusCode() != http.StatusOK:
			diags.AddError(
				"Unable to get database details while waiting for state",
				fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
			)
			return nil, diags
		case apiResp.JSON200 == nil:
			diags.AddError(
				"Unable to get database details while waiting for state",
				"API returned status code 200 without a database response body",
			)
			return nil, diags
		case apiResp.JSON200.State == nil:
			diags.AddError(
				"Unable to get database details while waiting for state",
				"API returned a database response without a state",
			)
			return nil, diags
		case *apiResp.JSON200.State == databaseStateError:
			diags.AddError(
				"Managed database entered error state",
				databaseStateErrorDetails(apiResp.JSON200.StateError),
			)
			return nil, diags
		case match(*apiResp.JSON200.State):
			return apiResp.JSON200, diags
		default:
			serverErrors = 0
		}

		select {
		case <-ctx.Done():
			diags.AddError(
				"Context cancelled while waiting for database state",
				utils.ErrorDiagnosticDetail(ctx.Err()),
			)
			return nil, diags
		case <-ticker.C:
		}
	}
}

func waitForDatabaseState(ctx context.Context, client *v9.ClientWithResponses, id, desiredState string) (*v9.DatabaseServiceInformationResponse, diag.Diagnostics) {
	db, diags := pollDatabase(ctx, client, id, func(state string) bool {
		return state == desiredState
	}, false)
	if diags.HasError() {
		return nil, diags
	}
	if db == nil {
		diags.AddError(
			fmt.Sprintf("Error while waiting for database to be in %s state", desiredState),
			"Database was deleted while waiting",
		)
	}
	return db, diags
}

func waitForPoweredState(ctx context.Context, client *v9.ClientWithResponses, id string, powered bool) (*v9.DatabaseServiceInformationResponse, diag.Diagnostics) {
	expectedState := databaseStateRunning
	if !powered {
		expectedState = databaseStateStopped
	}

	return waitForDatabaseState(ctx, client, id, expectedState)
}

func deleteDatabase(ctx context.Context, client *v9.ClientWithResponses, id string) (diags diag.Diagnostics) {
	serviceUUID, err := uuid.Parse(id)
	if err != nil {
		diags.AddError(
			"Unable to delete managed database",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}

	apiResp, err := client.DeleteDatabaseWithResponse(ctx, serviceUUID)
	if err != nil {
		diags.AddError(
			"Unable to delete managed database",
			utils.ErrorDiagnosticDetail(err),
		)
		return diags
	}
	if apiResp.StatusCode() != http.StatusNoContent {
		diags.AddError(
			"Unable to delete managed database",
			fmt.Sprintf("Unexpected API status code %d: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return diags
	}

	_, d := pollDatabase(ctx, client, id, func(string) bool { return false }, true)
	diags.Append(d...)
	return diags
}

func waitForNonPendingState(ctx context.Context, client *v9.ClientWithResponses, id string) diag.Diagnostics {
	_, diags := pollDatabase(ctx, client, id, func(state string) bool {
		return state != databaseStatePending
	}, false)
	return diags
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
