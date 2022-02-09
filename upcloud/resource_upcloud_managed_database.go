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

	"github.com/UpCloudLtd/terraform-provider-upcloud/upcloud/upcloudschema"
)

const (
	managedDatabaseTypePostgreSQL = upcloud.ManagedDatabaseServiceTypePostgreSQL
	managedDatabaseTypeMySQL      = upcloud.ManagedDatabaseServiceTypeMySQL
)

var resourceUpcloudManagedDatabaseModifiableStates = []upcloud.ManagedDatabaseState{
	upcloud.ManagedDatabaseStateRunning,
	upcloud.ManagedDatabaseState("rebalancing"),
}

func resourceUpCloudManagedDatabasePostgreSQL() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents PostgreSQL managed database",
		CreateContext: resourceUpCloudManagedDatabaseCreate(managedDatabaseTypePostgreSQL),
		ReadContext:   resourceUpCloudManagedDatabaseRead,
		UpdateContext: resourceUpCloudManagedDatabaseUpdate,
		DeleteContext: resourceUpCloudManagedDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: buildSchemaForUpCloudManagedDatabase(schemaUpcloudManagedDatabaseEngine(managedDatabaseTypePostgreSQL)),
	}
}

func resourceUpCloudManagedDatabaseMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents MySQL managed database",
		CreateContext: resourceUpCloudManagedDatabaseCreate(managedDatabaseTypeMySQL),
		ReadContext:   resourceUpCloudManagedDatabaseRead,
		UpdateContext: resourceUpCloudManagedDatabaseUpdate,
		DeleteContext: resourceUpCloudManagedDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: buildSchemaForUpCloudManagedDatabase(schemaUpcloudManagedDatabaseEngine(managedDatabaseTypeMySQL)),
	}
}

func schemaUpcloudManagedDatabaseEngine(serviceType upcloud.ManagedDatabaseServiceType) map[string]*schema.Schema {
	switch serviceType {
	case managedDatabaseTypePostgreSQL:
		return map[string]*schema.Schema{
			"primary_database": {
				Description: "Primary database name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "Database Engine properties for PostgreSQL",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: buildManagedDatabasePropertiesSchema(serviceType)},
			},
			"sslmode": {
				Description: "SSL Connection Mode for PostgreSQL",
				Type:        schema.TypeString,
				Computed:    true,
			},
		}
	case managedDatabaseTypeMySQL:
		return map[string]*schema.Schema{
			"primary_database": {
				Description: "Primary database name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"properties": {
				Description: "Database Engine properties for MySQL",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: buildManagedDatabasePropertiesSchema(serviceType)},
			},
		}
	}
	return nil
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
		"components": {
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
		},
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
		"node_states": {
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
		},
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
	}
}

func buildSchemaForUpCloudManagedDatabase(engineSchema map[string]*schema.Schema) map[string]*schema.Schema {
	serviceSchema := schemaUpCloudManagedDatabaseCommon()
	for k, v := range engineSchema {
		serviceSchema[k] = v
	}
	return serviceSchema
}

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
			req.Properties = buildManagedDatabasePropertiesFromResourceData(
				d,
				upcloudschema.ManagedDatabaseServicePropertiesSchema(serviceType),
				"properties", "0",
			)
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

func resourceUpCloudManagedDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	return copyManagedDatabaseDetailsToResourceData(d, details)
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
			req.Properties = buildManagedDatabasePropertiesFromResourceData(
				d,
				upcloudschema.ManagedDatabaseServicePropertiesSchema(upcloud.ManagedDatabaseServiceType(d.Get("type").(string))),
				"properties", "0",
			)
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

func buildManagedDatabaseSubResourceID(serviceID, subResourceID string) string {
	return fmt.Sprintf("%s/%s", serviceID, subResourceID)
}

func splitManagedDatabaseSubResourceID(packed string) (serviceID string, subResourceID string) {
	parts := strings.SplitN(packed, "/", 2)
	serviceID = parts[0]
	if len(parts) > 1 {
		subResourceID = parts[1]
	}
	return serviceID, subResourceID
}

func copyManagedDatabaseDetailsToResourceData(d *schema.ResourceData, details *upcloud.ManagedDatabase) diag.Diagnostics {
	convertComponents := func(details *upcloud.ManagedDatabase) (r []map[string]interface{}) {
		for _, comp := range details.Components {
			r = append(r, map[string]interface{}{
				"component": comp.Component,
				"host":      comp.Host,
				"port":      comp.Port,
				"route":     comp.Route,
				"usage":     comp.Usage,
			})
		}
		return r
	}
	convertNodes := func(details *upcloud.ManagedDatabase) (r []map[string]interface{}) {
		for _, node := range details.NodeStates {
			r = append(r, map[string]interface{}{
				"name":  node.Name,
				"role":  node.Role,
				"state": node.State,
			})
		}
		return r
	}
	type sf struct {
		name string
		val  interface{}
	}
	setFields := []sf{
		{name: "name", val: details.Name},
		{name: "components", val: convertComponents(details)},
		{name: "maintenance_window_dow", val: details.Maintenance.DayOfWeek},
		{name: "maintenance_window_time", val: details.Maintenance.Time},
		{name: "node_states", val: convertNodes(details)},
		{name: "plan", val: details.Plan},
		{name: "service_uri", val: details.ServiceURI},
		{name: "service_host", val: details.ServiceURIParams.Host},
		{name: "service_port", val: details.ServiceURIParams.Port},
		{name: "service_username", val: details.ServiceURIParams.User},
		{name: "service_password", val: details.ServiceURIParams.Password},
		{name: "state", val: string(details.State)},
		{name: "title", val: details.Title},
		{name: "zone", val: details.Zone},
		{name: "powered", val: details.Powered},
	}

	switch d.Get("type") {
	case "pg":
		setFields = append(setFields,
			sf{name: "primary_database", val: details.ServiceURIParams.DatabaseName},
			sf{name: "sslmode", val: details.ServiceURIParams.SSLMode})
	case "mysql":
		setFields = append(setFields,
			sf{name: "primary_database", val: details.ServiceURIParams.DatabaseName})
	}

	for _, sf := range setFields {
		if err := d.Set(sf.name, sf.val); err != nil {
			return diag.FromErr(err)
		}
	}

	if len(details.Properties) > 0 {
		props := make(map[string]interface{})
		for k, v := range details.Properties {
			props[string(k)] = v
		}

		newProps, err := buildManagedDatabasePropertiesResourceDataFromAPIProperties(props,
			upcloudschema.ManagedDatabaseServicePropertiesSchema(details.Type))
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("properties", []interface{}{newProps}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
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

func buildManagedDatabasePropertiesResourceDataFromAPIProperties(
	apiProps map[string]interface{},
	propertiesSchema map[string]interface{},
) (map[string]interface{}, error) {
	r := make(map[string]interface{})
	for k, iv := range apiProps {
		propertySchema := propertiesSchema[k].(map[string]interface{})
		var newValue interface{}
		switch propertySchema["type"] {
		//nolint // not really necessary to make these constants
		case "object":
			if _, ok := iv.(map[string]interface{}); !ok {
				return nil, fmt.Errorf("invalid api response for property %q (not an object)", k)
			}
			objectProperties := propertySchema["properties"].(map[string]interface{})
			subProps, err := buildManagedDatabasePropertiesResourceDataFromAPIProperties(iv.(map[string]interface{}), objectProperties)
			if err != nil {
				return nil, err
			}
			newValue = []interface{}{subProps}
		default:
			newValue = iv
		}
		r[k] = newValue
	}
	return r, nil
}

func buildManagedDatabasePropertiesFromResourceData(
	d *schema.ResourceData,
	propertiesSchema map[string]interface{},
	keyPath ...string,
) map[upcloud.ManagedDatabasePropertyKey]interface{} {
	resourceProps := d.Get(strings.Join(keyPath, ".")).(map[string]interface{})
	r := make(map[upcloud.ManagedDatabasePropertyKey]interface{})
	for k, iv := range resourceProps {
		if !d.HasChange(strings.Join(append(keyPath, k), ".")) {
			continue
		}
		propertySchema := propertiesSchema[k].(map[string]interface{})
		var setValue interface{}
		switch v := iv.(type) {
		case []interface{}:
			if propertySchema["type"] == "object" {
				setValue = make(map[upcloud.ManagedDatabasePropertyKey]interface{})
				if len(v) > 0 {
					objectProperties := propertySchema["properties"].(map[string]interface{})
					setValue = buildManagedDatabasePropertiesFromResourceData(d, objectProperties, append(keyPath, k, "0")...)
				}
			} else {
				setValue = v
			}
		default:
			setValue = v
		}
		r[upcloud.ManagedDatabasePropertyKey(k)] = setValue
	}
	return r
}

func composeManagedDatabasePropertiesOverrideChain(fns ...upcloudschema.FnGenerateTerraformSchemaOverride) upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		for _, fn := range fns {
			fn(keyPath, proposedSchema, source)
		}
	}
}

func overrideManagedDatabasePropertiesAllOptional() upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		proposedSchema.Optional = true
	}
}

func overrideManagedDatabasePropertiesComputed() upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		if proposedSchema.Default == nil && proposedSchema.DefaultFunc == nil {
			proposedSchema.Computed = true
		}
	}
}
func overrideManagedDatabasePropertiesDefaults() upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		def, ok := source["default"]
		if !ok {
			return
		}
		switch source["type"] {
		case "object", "array":
			return
		case "integer":
			def = int(def.(float64))
		default:
		}
		proposedSchema.Default = def
	}
}

func overrideManagedDatabasePropertiesIPFilter() upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		if len(keyPath) > 0 && keyPath[len(keyPath)-1] == "ip_filter" {
			prev := proposedSchema.DiffSuppressFunc
			proposedSchema.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
				if prev != nil && prev(k, old, new, d) {
					return true
				}
				if strings.TrimSuffix(old, "/32") == strings.TrimSuffix(new, "/32") {
					return true
				}
				return false
			}
		}
	}
}

func overrideManagedDatabasePropertiesMarkSensitive() upcloudschema.FnGenerateTerraformSchemaOverride {
	patterns := []string{"password"}
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		for _, pat := range patterns {
			if strings.Contains(keyPath[len(keyPath)-1], pat) {
				proposedSchema.Sensitive = true
			}
		}
	}
}

func overrideManagedDatabasePropertiesCreateOnly() upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		if v, ok := source["createOnly"].(bool); ok && v {
			proposedSchema.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
				return d.Id() != ""
			}
		}
	}
}

func overrideManagedDatabasePropertiesIgnoreClearing() upcloudschema.FnGenerateTerraformSchemaOverride {
	return func(keyPath []string, proposedSchema *schema.Schema, source map[string]interface{}) {
		prev := proposedSchema.DiffSuppressFunc
		proposedSchema.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
			if prev != nil && prev(k, old, new, d) {
				return true
			}
			if proposedSchema.Default != nil || proposedSchema.DefaultFunc != nil {
				return false
			}
			if old != "" && new == "" {
				log.Printf("[DEBUG] ignoring diff for %s (schema=%+v source=%+v)", k, proposedSchema, source)
				return true
			}

			return false
		}
	}
}

func buildManagedDatabasePropertiesSchema(serviceType upcloud.ManagedDatabaseServiceType) map[string]*schema.Schema {
	jsonSchema := upcloudschema.ManagedDatabaseServicePropertiesSchema(serviceType)
	overrides := composeManagedDatabasePropertiesOverrideChain(
		overrideManagedDatabasePropertiesAllOptional(),
		overrideManagedDatabasePropertiesIPFilter(),
		overrideManagedDatabasePropertiesMarkSensitive(),
		overrideManagedDatabasePropertiesDefaults(),
		overrideManagedDatabasePropertiesCreateOnly(),
		overrideManagedDatabasePropertiesComputed(),
		overrideManagedDatabasePropertiesIgnoreClearing(),
	)
	return upcloudschema.GenerateTerraformSchemaFromJSONSchema(jsonSchema, overrides)
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
