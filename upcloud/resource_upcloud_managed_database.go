package upcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	managedDatabaseTypePostgreSQL = upcloud.ManagedDatabaseServiceTypePostgreSQL
	managedDatabaseTypeMySQL      = upcloud.ManagedDatabaseServiceTypeMySQL
)

func resourceUpCloudManagedDatabaseCreate(serviceType upcloud.ManagedDatabaseServiceType) schema.CreateContextFunc {
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

		details, err := client.CreateManagedDatabase(&req)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(details.UUID)

		log.Printf("[INFO] managed database %v (%v) created", details.UUID, d.Get("name"))

		if err = waitManagedDatabaseFullyCreated(ctx, client, details); err != nil {
			d := resourceUpCloudManagedDatabaseRead(ctx, d, meta)
			d = append(d, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
			return d
		}

		if !d.Get("powered").(bool) {
			_, err := client.ShutdownManagedDatabase(&request.ShutdownManagedDatabaseRequest{UUID: d.Id()})
			if err != nil {
				return diag.FromErr(err)
			}
			log.Printf("[INFO] managed database %v (%v) is powered off", d.Id(), d.Get("name"))
		}

		if err = waitServiceNameToPropagate(ctx, details.ServiceURIParams.Host); err != nil {
			// return warning if DNS name is not yet available
			d := resourceUpCloudManagedDatabaseRead(ctx, d, meta)
			d = append(d, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
			return d
		}

		return resourceUpCloudManagedDatabaseRead(ctx, d, meta)
	}
}

func resourceUpCloudManagedDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var err error
	client := meta.(*service.Service)
	req := request.GetManagedDatabaseRequest{UUID: d.Id()}
	details, err := client.GetManagedDatabase(&req)
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == upcloudDatabaseNotFoundErrorCode {
			var diags diag.Diagnostics
			diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("name").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] managed database %v (%v) read", d.Id(), d.Get("name"))

	if d.Get("type").(string) == string(managedDatabaseTypePostgreSQL) {
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

func resourceUpCloudManagedDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

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
			req.Properties = buildManagedDatabasePropertiesRequestFromResourceData(d)
		}

		_, err := client.ModifyManagedDatabase(&req)
		if err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[INFO] managed database %v (%v) updated", d.Id(), d.Get("name"))
	}

	if d.HasChange("powered") {
		if d.Get("powered").(bool) {
			_, err := client.StartManagedDatabase(&request.StartManagedDatabaseRequest{UUID: d.Id()})
			if err != nil {
				return diag.FromErr(err)
			}
			log.Printf("[INFO] managed database %v (%v) is powered on", d.Id(), d.Get("name"))
		} else {
			_, err := client.ShutdownManagedDatabase(&request.ShutdownManagedDatabaseRequest{UUID: d.Id()})
			if err != nil {
				return diag.FromErr(err)
			}
			log.Printf("[INFO] managed database %v (%v) is powered off", d.Id(), d.Get("name"))
		}
	}

	return resourceUpCloudManagedDatabaseRead(ctx, d, meta)
}

func resourceUpCloudManagedDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	req := request.DeleteManagedDatabaseRequest{UUID: d.Id()}
	if err := client.DeleteManagedDatabase(&req); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] managed database %v (%v) deleted", req.UUID, d.Get("name"))

	return nil
}

func buildManagedDatabasePropertiesRequestFromResourceData(d *schema.ResourceData) request.ManagedDatabasePropertiesRequest {
	resourceProps := d.Get("properties.0").(map[string]interface{})
	r := make(map[upcloud.ManagedDatabasePropertyKey]interface{})
	for field, value := range resourceProps {
		if !d.HasChange(fmt.Sprintf("properties.0.%s", field)) {
			continue
		}
		switch field {
		case "migration", "pglookout", "timescaledb", "pgbouncer":
			// convert resource data list of objects into API objects
			if listValue, ok := value.([]interface{}); ok && len(listValue) == 1 {
				r[upcloud.ManagedDatabasePropertyKey(field)] = listValue[0]
			}
		case "pg_stat_statements_track", "pg_partman_bgw_role", "pg_partman_bgw_interval":
			// with these fields last underscore needs to be converted to dot e.g. pg_stat_statements_track -> pg_stat_statements.track
			c := strings.Split(field, "_")
			r[upcloud.ManagedDatabasePropertyKey(fmt.Sprintf("%s.%s", strings.Join(c[:len(c)-1], "_"), c[len(c)-1]))] = value
		default:
			r[upcloud.ManagedDatabasePropertyKey(field)] = value
		}
	}
	return r
}

func buildManagedDatabaseResourceDataProperties(db *upcloud.ManagedDatabase, d *schema.ResourceData) map[string]interface{} {
	props := d.Get("properties.0").(map[string]interface{})
	for key, iv := range db.Properties {
		switch key {
		case "migration", "pglookout", "timescaledb", "pgbouncer":
			// convert API objects into list of objects
			if m, ok := iv.(map[string]interface{}); ok {
				props[string(key)] = []map[string]interface{}{m}
			}
		case "pg_stat_statements.track", "pg_partman_bgw.role", "pg_partman_bgw.interval":
			// with these fields last dot needs to be converted to underscore e.g. pg_stat_statements.track -> pg_stat_statements_track
			props[(strings.Replace(string(key), ".", "_", 1))] = iv
		default:
			props[string(key)] = iv
		}
	}
	return props
}

func schemaUpCloudManagedDatabaseCommon() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description:  "Name of the service. The name is used as a prefix for the logical hostname. Must be unique within an account",
			Type:         schema.TypeString,
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(3, 30),
		},
		"components": schemaUpCloudManagedDatabaseComponents(),
		"maintenance_window_dow": {
			Description:      "Maintenance window day of week. Lower case weekday name (monday, tuesday, ...)",
			Type:             schema.TypeString,
			Computed:         true,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}, false)),
		},
		"maintenance_window_time": {
			Description: "Maintenance window UTC time in hh:mm:ss format",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
				if _, err := time.Parse("03:04:05", i.(string)); err != nil {
					return diag.FromErr(fmt.Errorf("invalid time"))
				}
				return nil
			},
		},
		"node_states": schemaUpCloudManagedDatabaseNodeStates(),
		"plan": {
			Description: "Service plan to use. This determines how much resources the instance will have",
			Type:        schema.TypeString,
			Required:    true,
		},
		"powered": {
			Description: "The administrative power state of the service",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"service_uri": {
			Description: "URI to the service instance",
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
		},
		"service_host": {
			Description: "Hostname to the service instance",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_port": {
			Description: "Port to the service instance",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_username": {
			Description: "Primary username to the service instance",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_password": {
			Description: "Primary username's password to the service instance",
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
		},
		"state": {
			Description: "State of the service",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"title": {
			Description:  "Title of a managed database instance",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(0, 255),
		},
		"type": {
			Description: "Type of the service",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"zone": {
			Description: "Zone where the instance resides",
			Type:        schema.TypeString,
			Required:    true,
		},
		"primary_database": {
			Description: "Primary database name",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func schemaUpCloudManagedDatabaseCommonProperties() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"admin_password": {
			Type:             schema.TypeString,
			Description:      "Custom password for admin user. Defaults to random string. This must be set only when a new service is being created.",
			Optional:         true,
			Computed:         true,
			Sensitive:        true,
			DiffSuppressFunc: diffSuppressCreateOnlyProperty,
		},
		"admin_username": {
			Type:             schema.TypeString,
			Description:      "Custom username for admin user. This must be set only when a new service is being created.",
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressCreateOnlyProperty,
		},
		"automatic_utility_network_ip_filter": {
			Type:        schema.TypeBool,
			Description: "Automatic utility network IP Filter",
			Optional:    true,
			Default:     true,
		},
		"backup_hour": {
			Type:             schema.TypeInt,
			Description:      "The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 23)),
		},
		"backup_minute": {
			Type:             schema.TypeInt,
			Description:      "The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 59)),
		},
		"ip_filter": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "IP filter",
			Optional:    true,
			Computed:    true,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return strings.TrimSuffix(old, "/32") == strings.TrimSuffix(new, "/32")
			},
		},
		"public_access": {
			Type:        schema.TypeBool,
			Description: "Public Access",
			Optional:    true,
			Default:     false,
		},
		"migration": {
			Description: "Migrate data from existing server",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem:        &schema.Resource{Schema: schemaUpCloudManagedDatabaseMigration()},
		},
	}
}

func schemaUpCloudManagedDatabaseComponents() *schema.Schema {
	return &schema.Schema{
		Description: "Service component information",
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"component": {
					Description: "Type of the component",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"host": {
					Description: "Hostname of the component",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"port": {
					Description: "Port number of the component",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"route": {
					Description: "Component network route type",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"usage": {
					Description: "Usage of the component",
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}
}

func schemaUpCloudManagedDatabaseNodeStates() *schema.Schema {
	return &schema.Schema{
		Description: "Information about nodes providing the managed service",
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "Name plus a node iteration",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"role": {
					Description: "Role of the node",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"state": {
					Description: "State of the node",
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}
}

func schemaUpCloudManagedDatabaseMigration() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"dbname": {
			Type:        schema.TypeString,
			Description: "Database name for bootstrapping the initial connection",
			Optional:    true,
			Computed:    true,
		},
		"host": {
			Type:        schema.TypeString,
			Description: "Hostname or IP address of the server where to migrate data from",
			Optional:    true,
			Computed:    true,
		},
		"ignore_dbs": {
			Type:        schema.TypeString,
			Description: "Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)",
			Optional:    true,
			Computed:    true,
		},
		"password": {
			Type:        schema.TypeString,
			Description: "Password for authentication with the server where to migrate data from",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
		},
		"port": {
			Type:        schema.TypeInt,
			Description: "Port number of the server where to migrate data from",
			Optional:    true,
			Computed:    true,
		},
		"ssl": {
			Type:        schema.TypeBool,
			Description: "The server where to migrate data from is secured with SSL",
			Optional:    true,
			Default:     true,
		},
		"username": {
			Type:        schema.TypeString,
			Description: "User name for authentication with the server where to migrate data from",
			Optional:    true,
			Computed:    true,
		},
	}
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
			if db, err = client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: db.UUID}); err != nil {
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

func isManagedDatabaseFullyCreated(db *upcloud.ManagedDatabase) bool {
	return db.State == upcloud.ManagedDatabaseStateRunning && len(db.Backups) > 0 && len(db.Users) > 0
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

func resourceUpCloudManagedDatabaseWaitState(
	ctx context.Context,
	id string,
	m interface{},
	timeout time.Duration,
	targetStates ...upcloud.ManagedDatabaseState,
) (*upcloud.ManagedDatabase, error) {
	client := m.(*service.Service)
	refresher := func() (result interface{}, state string, err error) {
		resp, err := client.GetManagedDatabase(&request.GetManagedDatabaseRequest{UUID: id})
		if err != nil {
			return nil, "", err
		}
		return resp, string(resp.State), nil
	}
	res, state, err := refresher()
	if err != nil {
		return nil, err
	}
	if len(targetStates) == 0 {
		return res.(*upcloud.ManagedDatabase), nil
	}
	for _, targetState := range targetStates {
		if upcloud.ManagedDatabaseState(state) == targetState {
			return res.(*upcloud.ManagedDatabase), nil
		}
	}
	states := make([]string, 0, len(targetStates))
	for _, targetState := range targetStates {
		states = append(states, string(targetState))
	}
	waiter := resource.StateChangeConf{
		Delay:      1 * time.Second,
		Refresh:    refresher,
		Target:     states,
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}
	res, err = waiter.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}
	return res.(*upcloud.ManagedDatabase), nil
}

func splitManagedDatabaseSubResourceID(packed string) (serviceID string, subResourceID string) {
	parts := strings.SplitN(packed, "/", 2)
	serviceID = parts[0]
	if len(parts) > 1 {
		subResourceID = parts[1]
	}
	return serviceID, subResourceID
}

func buildManagedDatabaseSubResourceID(serviceID, subResourceID string) string {
	return fmt.Sprintf("%s/%s", serviceID, subResourceID)
}

var resourceUpcloudManagedDatabaseModifiableStates = []upcloud.ManagedDatabaseState{
	upcloud.ManagedDatabaseStateRunning,
	upcloud.ManagedDatabaseState("rebalancing"),
}

func diffSuppressCreateOnlyProperty(k, old, new string, d *schema.ResourceData) bool {
	return d.Id() != ""
}
