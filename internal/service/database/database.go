package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceUpcloudManagedDatabaseModifiableStates = []upcloud.ManagedDatabaseState{
	upcloud.ManagedDatabaseStateRunning,
	upcloud.ManagedDatabaseState("rebalancing"),
}

func resourceDatabaseCreate(serviceType upcloud.ManagedDatabaseServiceType) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*service.Service)

		if err := d.Set("type", string(serviceType)); err != nil {
			return diag.FromErr(err)
		}

		req := request.CreateManagedDatabaseRequest{
			HostNamePrefix: d.Get("name").(string),
			Plan:           d.Get("plan").(string),
			Title:          d.Get("title").(string),
			Type:           serviceType,
			Zone:           d.Get("zone").(string),
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

		details, err := client.CreateManagedDatabase(ctx, &req)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(details.UUID)

		tflog.Info(ctx, "managed database created", map[string]interface{}{"uuid": details.UUID, "name": d.Get("name")})

		if err = waitManagedDatabaseFullyCreated(ctx, client, details); err != nil {
			d := resourceDatabaseRead(ctx, d, meta)
			d = append(d, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
			return d
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

		return resourceDatabaseRead(ctx, d, meta)
	}
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

	if d.Get("type").(string) == string(upcloud.ManagedDatabaseServiceTypePostgreSQL) {
		if err := d.Set("sslmode", details.ServiceURIParams.SSLMode); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := resourceUpCloudManagedDatabaseSetCommonState(d, details); err != nil {
		return diag.FromErr(err)
	}

	if len(details.Properties) > 0 {
		if err := d.Set("properties", []map[string]interface{}{buildManagedDatabaseResourceDataProperties(details, d)}); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)
	diags := diag.Diagnostics{}

	if d.HasChanges("plan", "title", "zone",
		"maintenance_window_dow", "maintenance_window_time", "properties.0") {
		req := request.ModifyManagedDatabaseRequest{UUID: d.Id()}
		req.Plan = d.Get("plan").(string)
		req.Title = d.Get("title").(string)
		req.Zone = d.Get("zone").(string)
		if d.HasChange("maintenance_window_dow") || d.HasChange("maintenance_window_time") {
			req.Maintenance.DayOfWeek = d.Get("maintenance_window_dow").(string)
			req.Maintenance.Time = d.Get("maintenance_window_time").(string)
		}

		if d.HasChange("properties.0") {
			props := buildManagedDatabasePropertiesRequestFromResourceData(d)

			// Always delete version if it exists; versions are updated via separate endpoint
			delete(props, "version")
			req.Properties = props
		}

		_, err := client.ModifyManagedDatabase(ctx, &req)
		if err != nil {
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "managed database updated", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})
	}

	if d.HasChange("powered") {
		if d.Get("powered").(bool) {
			_, err := client.StartManagedDatabase(ctx, &request.StartManagedDatabaseRequest{UUID: d.Id()})
			if err != nil {
				return diag.FromErr(err)
			}
			tflog.Info(ctx, "managed database is powered on", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})

			// Attempt to upgrade version after the database was powered on
			if d.HasChange("properties.0.version") {
				// Upgrade is only allowed when database is in "Running" state, so we have to wait for that after powering it on
				_, err := resourceUpCloudManagedDatabaseWaitState(ctx, d.Id(), client, time.Minute*15, upcloud.ManagedDatabaseStateRunning)

				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  fmt.Sprintf("Upgrading Managed Database %s(%s) version failed; reached timeout when waiting for running state", d.Id(), d.Get("name")),
						Detail:   err.Error(),
					})
				} else {
					diags = append(diags, updateDatabaseVersion(ctx, d, client)...)
				}
			}
		} else {
			// Attempt to upgrade version before database is powered off
			if d.HasChange("properties.0.version") {
				diags = append(diags, updateDatabaseVersion(ctx, d, client)...)
			}

			_, err := client.ShutdownManagedDatabase(ctx, &request.ShutdownManagedDatabaseRequest{UUID: d.Id()})
			if err != nil {
				return diag.FromErr(err)
			}
			tflog.Info(ctx, "managed database is powered off", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})
		}
	} else {
		// If powered state was not chaged, just attempt to upgrade version
		if d.HasChange("properties.0.version") {
			if !d.Get("powered").(bool) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("Version upgrade for Managed Database %s(%s) skipped", d.Id(), d.Get("name")),
					Detail:   "Cannot upgrade version for Managed Database when it is powered off",
				})
			} else {
				diags = append(diags, updateDatabaseVersion(ctx, d, client)...)
			}
		}
	}

	diags = append(diags, resourceDatabaseRead(ctx, d, meta)...)
	return diags
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.DeleteManagedDatabaseRequest{UUID: d.Id()}
	if err := client.DeleteManagedDatabase(ctx, &req); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "managed database deleted", map[string]interface{}{"uuid": d.Id(), "name": d.Get("name")})

	return nil
}

func buildManagedDatabasePropertiesRequestFromResourceData(d *schema.ResourceData) request.ManagedDatabasePropertiesRequest {
	resourceProps := d.Get("properties.0").(map[string]interface{})
	r := make(map[upcloud.ManagedDatabasePropertyKey]interface{})
	for field := range resourceProps {
		key := fmt.Sprintf("properties.0.%s", field)
		value, isNotZero := d.GetOk(key)
		// It seems to be hard to detect changes in boolen fields in some scenarios.
		// E.g. if boolean field is optional and has default value true, but is initially set as false in config
		// then it's interpreted as boolean zero value and no change is detected.
		//
		// This might need more thinking, but for now, exclude field from the request if
		// it's not boolean, has type's "zero" value and value hasn't changed.
		if _, isBool := value.(bool); !isBool && !isNotZero && !d.HasChange(key) {
			continue
		}
		switch field {
		case "migration", "pglookout", "timescaledb", "pgbouncer":
			// convert resource data list of objects into API objects
			if listValue, ok := value.([]interface{}); ok && len(listValue) == 1 {
				r[upcloud.ManagedDatabasePropertyKey(field)] = listValue[0]
			}
		default:
			r[upcloud.ManagedDatabasePropertyKey(field)] = value
		}
	}
	return r
}

func buildManagedDatabaseResourceDataProperties(db *upcloud.ManagedDatabase, d *schema.ResourceData) map[string]interface{} {
	props := d.Get("properties.0").(map[string]interface{})
	for key, iv := range db.Properties {
		if _, ok := props[string(key)]; !ok {
			continue
		}
		switch key {
		case "migration", "pglookout", "timescaledb", "pgbouncer":
			// convert API objects into list of objects
			if m, ok := iv.(map[string]interface{}); ok {
				props[string(key)] = []map[string]interface{}{m}
			}
		default:
			props[string(key)] = iv
		}
	}
	return props
}

func resourceUpCloudManagedDatabaseSetCommonState(d *schema.ResourceData, details *upcloud.ManagedDatabase) error {
	var nodeStates, components []map[string]interface{}
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

	for _, node := range details.NodeStates {
		nodeStates = append(nodeStates, map[string]interface{}{
			"name":  node.Name,
			"role":  node.Role,
			"state": node.State,
		})
	}

	if err = d.Set("name", details.Name); err != nil {
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
	if err = d.Set("title", details.Title); err != nil {
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

func waitManagedDatabaseFullyCreated(ctx context.Context, client *service.Service, db *upcloud.ManagedDatabase) error {
	const maxRetries int = 100
	var err error
	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if db, err = client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: db.UUID}); err != nil {
				return err
			}
			if isManagedDatabaseFullyCreated(db) {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
	return errors.New("max retries reached while waiting for managed database instance to be created")
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

func updateDatabaseVersion(ctx context.Context, d *schema.ResourceData, client *service.Service) diag.Diagnostics {
	diags := diag.Diagnostics{}

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
