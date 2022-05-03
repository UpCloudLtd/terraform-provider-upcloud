package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func schemaDatabaseCommon() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description:  "Name of the service. The name is used as a prefix for the logical hostname. Must be unique within an account",
			Type:         schema.TypeString,
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(3, 30),
		},
		"components": schemaDatabaseComponents(),
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
		"node_states": schemaDatabaseNodeStates(),
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

func schemaDatabaseCommonProperties() map[string]*schema.Schema {
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
			Elem:        &schema.Resource{Schema: schemaDatabaseMigration()},
		},
	}
}

func schemaDatabaseComponents() *schema.Schema {
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

func schemaDatabaseNodeStates() *schema.Schema {
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

func schemaDatabaseMigration() map[string]*schema.Schema {
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
