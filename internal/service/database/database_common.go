package database

import (
	"fmt"
	"regexp"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
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
		"labels":     utils.LabelsSchema("managed database"),
		"components": schemaDatabaseComponents(),
		"maintenance_window_dow": {
			Description:      "Maintenance window day of week. Lower case weekday name (monday, tuesday, ...)",
			Type:             schema.TypeString,
			Computed:         true,
			Optional:         true,
			RequiredWith:     []string{"maintenance_window_time"},
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}, false)),
		},
		"maintenance_window_time": {
			Description:  "Maintenance window UTC time in hh:mm:ss format",
			Type:         schema.TypeString,
			Computed:     true,
			Optional:     true,
			RequiredWith: []string{"maintenance_window_dow"},
			ValidateDiagFunc: func(i interface{}, _ cty.Path) diag.Diagnostics {
				if _, err := time.Parse("15:04:05", i.(string)); err != nil {
					return diag.FromErr(fmt.Errorf("maintenance_window_time format must be HH:MM:SS"))
				}
				return nil
			},
		},
		"network":     schemaDatabaseNetwork(),
		"node_states": schemaDatabaseNodeStates(),
		"plan": {
			Description: "Service plan to use. This determines how much resources the instance will have. You can list available plans with `upctl database plans <type>`.",
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
		"termination_protection": {
			Description: "If set to true, prevents the managed service from being powered off, or deleted.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"title": {
			Description:  "Title of a managed database instance",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(0, 255),
		},
		"type": {
			Description: "Type of the service",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"zone": {
			Description: "Zone where the instance resides, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"primary_database": {
			Description: "Primary database name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"additional_disk_space_gib": {
			Description: "Additional disk space in GiB",
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
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

func schemaDatabaseNetwork() *schema.Schema {
	return &schema.Schema{
		Description: "Private networks attached to the managed database",
		Type:        schema.TypeSet,
		Optional:    true,
		MaxItems:    8,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the network. Must be unique within the service.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.All(
						validation.StringLenBetween(0, 65),
						validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]+$"), ""),
					)),
				},
				"type": {
					Description: "The type of the network. Must be private.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice([]string{
							string(upcloud.LoadBalancerNetworkTypePrivate),
						}, false),
					),
				},
				"family": {
					Description: "Network family. Currently only `IPv4` is supported.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateDiagFunc: validation.ToDiagFunc(
						validation.StringInSlice([]string{
							string(upcloud.LoadBalancerAddressFamilyIPv4),
						}, false),
					),
				},
				"uuid": {
					Description: "Private network UUID. Must reside in the same zone as the database.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
	}
}

func networksFromResourceData(d *schema.ResourceData) []upcloud.ManagedDatabaseNetwork {
	req := make([]upcloud.ManagedDatabaseNetwork, 0)
	if networks, ok := d.GetOk("network"); ok {
		for _, network := range networks.(*schema.Set).List() {
			n := network.(map[string]interface{})
			uuid := n["uuid"].(string)
			r := upcloud.ManagedDatabaseNetwork{
				Name:   n["name"].(string),
				Type:   n["type"].(string),
				Family: n["family"].(string),
				UUID:   &uuid,
			}

			req = append(req, r)
		}
	}

	return req
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
					Optional:    true,
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
