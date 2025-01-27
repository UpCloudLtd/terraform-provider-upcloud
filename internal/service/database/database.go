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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func serviceDescription(dbType string) string {
	return fmt.Sprintf("This resource represents %s managed database. See UpCloud [Managed Databases](https://upcloud.com/products/managed-databases) product page for more details about the service.", dbType)
}

var resourceUpcloudManagedDatabaseModifiableStates = []upcloud.ManagedDatabaseState{
	upcloud.ManagedDatabaseStateRunning,
	upcloud.ManagedDatabaseState("rebalancing"),
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.Service)
	req := buildManagedDatabaseRequestFromResourceData(d)

	details, err := client.CreateManagedDatabase(ctx, &req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(details.UUID)

	tflog.Info(ctx, "managed database created", map[string]interface{}{"uuid": details.UUID, "name": d.Get("name")})

	if _, err = client.WaitForManagedDatabaseState(ctx, &request.WaitForManagedDatabaseStateRequest{UUID: details.UUID, DesiredState: upcloud.ManagedDatabaseStateRunning}); err != nil {
		return append(
			resourceDatabaseRead(ctx, d, meta),
			diag.FromErr(err)[0],
		)
	}

	if !d.Get("powered").(bool) {
		_, err := client.ShutdownManagedDatabase(ctx, &request.ShutdownManagedDatabaseRequest{UUID: d.Id()})
		if err != nil {
			return diag.FromErr(err)
		}
		tflog.Info(ctx, "managed database is powered off", map[string]interface{}{"uuid": details.UUID, "name": d.Get("name")})
	}

	if err = waitServiceNameToPropagate(ctx, details.ServiceURIParams.Host); err != nil {
		// return warning if DNS name is not yet available
		d := resourceDatabaseRead(ctx, d, meta)
		d = append(d, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  err.Error(),
		})
		return d
	}

	return diags
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var err error
	client := meta.(*service.Service)
	req := request.GetManagedDatabaseRequest{UUID: d.Id()}
	details, err := client.GetManagedDatabase(ctx, &req)
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	tflog.Debug(ctx, "managed database read", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})
	if details.Type == upcloud.ManagedDatabaseServiceTypePostgreSQL {
		if err := d.Set("sslmode", details.ServiceURIParams.SSLMode); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := resourceUpCloudManagedDatabaseSetCommonState(d, details); err != nil {
		return diag.FromErr(err)
	}
	if len(details.Properties) > 0 {
		if err := d.Set("properties", []map[string]interface{}{buildManagedDatabaseResourceDataProperties(d, details)}); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)
	diags := diag.Diagnostics{}

	if d.HasChanges("plan", "title", "termination_protection", "zone",
		"maintenance_window_dow", "maintenance_window_time", "properties.0", "network", "labels") {
		req := request.ModifyManagedDatabaseRequest{UUID: d.Id()}
		req.Plan = d.Get("plan").(string)
		req.Title = d.Get("title").(string)
		req.Zone = d.Get("zone").(string)
		if d.HasChanges("maintenance_window_dow", "maintenance_window_time") {
			req.Maintenance.DayOfWeek = d.Get("maintenance_window_dow").(string)
			req.Maintenance.Time = d.Get("maintenance_window_time").(string)
		}

		if d.HasChange("labels") {
			labels := utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{}))
			req.Labels = &labels
		}

		if d.HasChange("termination_protection") {
			terminationProtection := d.Get("termination_protection").(bool)
			req.TerminationProtection = &terminationProtection
		}

		if d.HasChange("properties.0") {
			props := buildManagedDatabasePropertiesRequestFromResourceData(d)

			// Always delete version if it exists; versions are updated via separate endpoint
			delete(props, "version")
			req.Properties = props
		}

		if d.HasChange("network") {
			networks := networksFromResourceData(d)
			req.Networks = &networks
		}

		_, err := client.ModifyManagedDatabase(ctx, &req)
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "managed database updated", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})
	}

	// Modify powered state if no PostgreSQL version update is requested
	if d.HasChange("powered") && !d.HasChange("properties.0.version") {
		return append(diags, resourceDatabasePoweredUpdate(ctx, d, client)...)
	}

	return diags
}

func resourceDatabasePoweredUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	client := meta.(*service.Service)

	var err error
	var msg string

	if d.Get("powered").(bool) {
		_, err = client.StartManagedDatabase(ctx, &request.StartManagedDatabaseRequest{UUID: d.Id()})
		msg = "managed database is powered on"
	} else {
		_, err = client.ShutdownManagedDatabase(ctx, &request.ShutdownManagedDatabaseRequest{UUID: d.Id()})
		msg = "managed database is powered off"
	}
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, msg, map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})

	return diags
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.DeleteManagedDatabaseRequest{UUID: d.Id()}
	if err := client.DeleteManagedDatabase(ctx, &req); err != nil {
		return diag.FromErr(err)
	}

	// Wait until database is deleted to be able to delete attached networks (if needed)
	if err := waitForDatabaseToBeDeleted(ctx, client, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "managed database deleted", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})

	return nil
}

func buildManagedDatabaseRequestFromResourceData(d *schema.ResourceData) request.CreateManagedDatabaseRequest {
	terminationProtection := d.Get("termination_protection").(bool)
	req := request.CreateManagedDatabaseRequest{
		HostNamePrefix:        d.Get("name").(string),
		Plan:                  d.Get("plan").(string),
		Title:                 d.Get("title").(string),
		TerminationProtection: &terminationProtection,
		Labels:                utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{})),
		Type:                  upcloud.ManagedDatabaseServiceType(d.Get("type").(string)),
		Zone:                  d.Get("zone").(string),
	}

	if d.HasChange("network") {
		req.Networks = networksFromResourceData(d)
	}

	if d.HasChange("properties.0") {
		req.Properties = buildManagedDatabasePropertiesRequestFromResourceData(d)
	}

	if d.HasChange("maintenance_window_dow") || d.HasChange("maintenance_window_time") {
		req.Maintenance = request.ManagedDatabaseMaintenanceTimeRequest{
			DayOfWeek: d.Get("maintenance_window_dow").(string),
			Time:      d.Get("maintenance_window_time").(string),
		}
	}

	return req
}

func getPropertiesValue(d *schema.ResourceData, key string) (interface{}, bool, bool) {
	value, isNotZero := d.GetOk(key)
	// It seems to be hard to detect changes in boolean fields in some scenarios.
	// E.g. if boolean field is optional and has default value true, but is initially set as false in config
	// then it's interpreted as boolean zero value and no change is detected.
	//
	// This might need more thinking, but for now, exclude field from the request if
	// it's not boolean, has type's "zero" value and value hasn't changed.
	_, isBool := value.(bool)
	hasChange := d.HasChange(key)
	shouldOmit := !isBool && !isNotZero && !hasChange

	return value, hasChange, shouldOmit
}

func buildManagedDatabasePropertiesRequestFromResourceData(d *schema.ResourceData) request.ManagedDatabasePropertiesRequest {
	resourceProps := d.Get("properties.0").(map[string]interface{})
	r := make(map[upcloud.ManagedDatabasePropertyKey]interface{})

	dbType := upcloud.ManagedDatabaseServiceType(d.Get("type").(string))
	propsInfo := properties.GetProperties(dbType)

	for field := range resourceProps {
		// skip properties that are not present in the propsInfo
		prop, ok := propsInfo[field]
		if !ok {
			continue
		}

		key := fmt.Sprintf("properties.0.%s", field)

		value, hasChange, shouldOmit := getPropertiesValue(d, key)

		if shouldOmit {
			continue
		}
		if prop.CreateOnly {
			if !hasChange {
				continue
			}
		}
		if properties.GetType(prop) == "object" {
			// convert resource data list of objects into API objects
			if listValue, ok := value.([]interface{}); ok && len(listValue) == 1 {
				// Do similar filtering for nested properties as is done for main level properties.
				stateObj := listValue[0].(map[string]interface{})
				reqObj := make(map[string]interface{})
				for k := range stateObj {
					value, _, shouldOmit := getPropertiesValue(d, fmt.Sprintf("%s.0.%s", key, k))
					reqKey := properties.GetKey(prop.Properties, k)
					if !shouldOmit && reqKey != "" {
						reqObj[reqKey] = value
					}
				}
				r[upcloud.ManagedDatabasePropertyKey(field)] = reqObj
			}
		} else {
			r[upcloud.ManagedDatabasePropertyKey(field)] = value
		}
	}
	return r
}

func buildManagedDatabaseResourceDataProperties(d *schema.ResourceData, db *upcloud.ManagedDatabase) map[string]interface{} {
	resourceProps := d.Get("properties.0").(map[string]interface{})
	propsInfo := properties.GetProperties(db.Type)

	for typedKey, value := range db.Properties {
		key := string(typedKey)

		// only use value from current state if it's a create-only property
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
				resourceProps[properties.SchemaKey(key)] = []map[string]interface{}{sMap}
			}
		} else {
			resourceProps[properties.SchemaKey(key)] = value
		}
	}

	// clean up removed properties that are not present in the propsInfo
	for key := range resourceProps {
		if _, ok := propsInfo[key]; !ok {
			delete(resourceProps, properties.SchemaKey(key))
		}
	}

	return resourceProps
}

func resourceUpCloudManagedDatabaseSetCommonState(d *schema.ResourceData, details *upcloud.ManagedDatabase) error {
	var components, networks, nodeStates []map[string]interface{}
	var err error

	for _, comp := range details.Components {
		components = append(components, map[string]interface{}{
			"component": comp.Component,
			"host":      comp.Host,
			"port":      comp.Port,
			"route":     comp.Route,
			"usage":     comp.Usage,
		})
	}

	for _, network := range details.Networks {
		networks = append(networks, map[string]interface{}{
			"family": network.Family,
			"name":   network.Name,
			"type":   network.Type,
			"uuid":   network.UUID,
		})
	}

	for _, node := range details.NodeStates {
		nodeState := map[string]interface{}{
			"name":  node.Name,
			"state": node.State,
		}
		if node.Role != "" {
			nodeState["role"] = node.Role
		}
		nodeStates = append(nodeStates, nodeState)
	}

	if err = d.Set("name", details.Name); err != nil {
		return err
	}
	if err := d.Set("labels", utils.LabelsSliceToMap(details.Labels)); err != nil {
		return err
	}
	if err = d.Set("components", components); err != nil {
		return err
	}
	if err = d.Set("maintenance_window_dow", details.Maintenance.DayOfWeek); err != nil {
		return err
	}
	if err = d.Set("maintenance_window_time", details.Maintenance.Time); err != nil {
		return err
	}
	if err = d.Set("network", networks); err != nil {
		return err
	}
	if err = d.Set("node_states", nodeStates); err != nil {
		return err
	}
	if err = d.Set("plan", details.Plan); err != nil {
		return err
	}
	if err = d.Set("service_uri", details.ServiceURI); err != nil {
		return err
	}
	if err = d.Set("service_host", details.ServiceURIParams.Host); err != nil {
		return err
	}
	if err = d.Set("service_port", details.ServiceURIParams.Port); err != nil {
		return err
	}
	if err = d.Set("service_username", details.ServiceURIParams.User); err != nil {
		return err
	}
	if err = d.Set("service_password", details.ServiceURIParams.Password); err != nil {
		return err
	}
	if err = d.Set("state", string(details.State)); err != nil {
		return err
	}
	if err = d.Set("termination_protection", details.TerminationProtection); err != nil {
		return err
	}
	if err = d.Set("title", details.Title); err != nil {
		return err
	}
	if err = d.Set("type", string(details.Type)); err != nil {
		return err
	}
	if err = d.Set("zone", details.Zone); err != nil {
		return err
	}
	if err = d.Set("powered", details.Powered); err != nil {
		return err
	}

	return d.Set("primary_database", details.ServiceURIParams.DatabaseName)
}

func getDatabaseDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	db, err := svc.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: id[0]})

	return map[string]interface{}{"resource": "database", "name": db.Name, "state": db.State}, err
}

func waitForDatabaseToBeDeleted(ctx context.Context, svc *service.Service, id string) error {
	return utils.WaitForResourceToBeDeleted(ctx, svc, getDatabaseDeleted, id)
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

func updateDatabaseVersion(ctx context.Context, d *schema.ResourceData, client *service.Service) (diags diag.Diagnostics) {
	// Cannot proceed with upgrade if powered off
	if !d.HasChange("powered") && !d.Get("powered").(bool) {
		return append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Version upgrade for Managed Database %s(%s) skipped", d.Id(), d.Get("name")),
			Detail:   "Cannot upgrade version for Managed Database when it is powered off",
		})
	}

	// Attempt to upgrade version after database is powered on
	// Upgrade is only allowed when database is in "Running" state, so we have to wait for that after powering it on
	if d.HasChange("powered") && d.Get("powered").(bool) {
		_, err := resourceUpCloudManagedDatabaseWaitState(ctx, d.Id(), client, time.Minute*15, upcloud.ManagedDatabaseStateRunning)
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Upgrading Managed Database %s(%s) version failed; reached timeout when waiting for running state", d.Id(), d.Get("name")),
				Detail:   err.Error(),
			})
		}
	}

	_, err := client.UpgradeManagedDatabaseVersion(ctx, &request.UpgradeManagedDatabaseVersionRequest{
		UUID:          d.Id(),
		TargetVersion: d.Get("properties.0.version").(string),
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Upgrading Managed Database %s(%s) version failed", d.Id(), d.Get("name")),
			Detail:   err.Error(),
		})
	}

	return diags
}
