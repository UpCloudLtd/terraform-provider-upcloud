package database

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/database/properties"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type databaseCommonModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Labels                 types.Map    `tfsdk:"labels"`
	Components             types.List   `tfsdk:"components"`
	MaintenanceWindowDow   types.String `tfsdk:"maintenance_window_dow"`
	MaintenanceWindowTime  types.String `tfsdk:"maintenance_window_time"`
	AdditionalDiskSpaceGiB types.Int64  `tfsdk:"additional_disk_space_gib"`
	Network                types.Set    `tfsdk:"network"`
	NodeStates             types.List   `tfsdk:"node_states"`
	Plan                   types.String `tfsdk:"plan"`
	Powered                types.Bool   `tfsdk:"powered"`
	ServiceURI             types.String `tfsdk:"service_uri"`
	ServiceHost            types.String `tfsdk:"service_host"`
	ServicePort            types.String `tfsdk:"service_port"`
	ServiceUsername        types.String `tfsdk:"service_username"`
	ServicePassword        types.String `tfsdk:"service_password"`
	State                  types.String `tfsdk:"state"`
	TerminationProtection  types.Bool   `tfsdk:"termination_protection"`
	Title                  types.String `tfsdk:"title"`
	Type                   types.String `tfsdk:"type"`
	Zone                   types.String `tfsdk:"zone"`
	PrimaryDatabase        types.String `tfsdk:"primary_database"`
	Properties             types.List   `tfsdk:"properties"`
}

type databaseComponentModel struct {
	Component types.String `tfsdk:"component"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int64  `tfsdk:"port"`
	Route     types.String `tfsdk:"route"`
	Usage     types.String `tfsdk:"usage"`
}

type databaseNetworkModel struct {
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Family types.String `tfsdk:"family"`
	UUID   types.String `tfsdk:"uuid"`
}

type databaseNodeStateModel struct {
	Name  types.String `tfsdk:"name"`
	Role  types.String `tfsdk:"role"`
	State types.String `tfsdk:"state"`
}

func defineCommonAttributesAndBlocks(s *schema.Schema, dbType upcloud.ManagedDatabaseServiceType) {
	planDescription := fmt.Sprintf("Service plan to use. This determines how much resources the instance will have. You can list available plans with `upctl database plans %s`.", dbType)
	additionalDiskDescription := "Additional disk space in GiB. Note that changes in additional disk space might require disk maintenance. This pending maintenance blocks some operations, such as version upgrades, until the maintenance is completed."
	if dbType == upcloud.ManagedDatabaseServiceTypeValkey {
		additionalDiskDescription = fmt.Sprintf("Not supported for `%s` databases. Should be left unconfigured.", dbType)
	}

	s.Attributes["id"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "UUID of the database.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	s.Attributes["name"] = schema.StringAttribute{
		MarkdownDescription: "Name of the service. The name is used as a prefix for the logical hostname. Must be unique within an account",
		Required:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			stringvalidator.LengthBetween(3, 30),
		},
	}
	s.Attributes["labels"] = utils.LabelsAttribute("database")
	s.Attributes["components"] = schema.ListNestedAttribute{
		MarkdownDescription: "Service component information",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"component": schema.StringAttribute{
					MarkdownDescription: "Component name.",
					Computed:            true,
				},
				"host": schema.StringAttribute{
					MarkdownDescription: "Hostname of the component",
					Computed:            true,
				},
				"port": schema.Int64Attribute{
					MarkdownDescription: "Port number of the component",
					Computed:            true,
				},
				"route": schema.StringAttribute{
					MarkdownDescription: "Component network route type",
					Computed:            true,
				},
				"usage": schema.StringAttribute{
					MarkdownDescription: "Usage of the component",
					Computed:            true,
				},
			},
		},
	}
	s.Attributes["maintenance_window_dow"] = schema.StringAttribute{
		MarkdownDescription: "Maintenance window day of week. Lower case weekday name (monday, tuesday, ...)",
		Computed:            true,
		Optional:            true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
			),
			stringvalidator.AlsoRequires(path.MatchRoot("maintenance_window_time")),
		},
	}
	s.Attributes["maintenance_window_time"] = schema.StringAttribute{
		MarkdownDescription: "Maintenance window UTC time in hh:mm:ss format",
		Computed:            true,
		Optional:            true,
		Validators: []validator.String{
			// TODO validate time format
			stringvalidator.AlsoRequires(path.MatchRoot("maintenance_window_dow")),
		},
	}
	s.Attributes["node_states"] = schema.ListNestedAttribute{
		MarkdownDescription: "Information about nodes providing the managed service",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Name plus a node iteration",
					Computed:            true,
				},
				"role": schema.StringAttribute{
					MarkdownDescription: "Role of the node",
					Computed:            true,
				},
				"state": schema.StringAttribute{
					MarkdownDescription: "Current state of the node",
					Computed:            true,
				},
			},
		},
	}
	s.Attributes["plan"] = schema.StringAttribute{
		MarkdownDescription: planDescription,
		Required:            true,
	}
	s.Attributes["powered"] = schema.BoolAttribute{
		MarkdownDescription: "The administrative power state of the service",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
	}
	s.Attributes["service_uri"] = schema.StringAttribute{
		MarkdownDescription: "URI to the service instance",
		Computed:            true,
		Sensitive:           true,
	}
	s.Attributes["service_host"] = schema.StringAttribute{
		MarkdownDescription: "Hostname to the service instance",
		Computed:            true,
	}
	s.Attributes["service_port"] = schema.StringAttribute{
		MarkdownDescription: "Port to the service instance",
		Computed:            true,
	}
	s.Attributes["service_username"] = schema.StringAttribute{
		MarkdownDescription: "Primary username to the service instance",
		Computed:            true,
	}
	s.Attributes["service_password"] = schema.StringAttribute{
		MarkdownDescription: "Primary password to the service instance",
		Computed:            true,
		Sensitive:           true,
	}
	s.Attributes["state"] = schema.StringAttribute{
		MarkdownDescription: "The current state of the service",
		Computed:            true,
	}
	s.Attributes["termination_protection"] = schema.BoolAttribute{
		MarkdownDescription: "If set to true, prevents the managed service from being powered off, or deleted.",
		Optional:            true,
		Default:             booldefault.StaticBool(false),
		Computed:            true,
	}
	s.Attributes["title"] = schema.StringAttribute{
		MarkdownDescription: "Title of the managed database instance",
		Required:            true,
	}
	s.Attributes["type"] = schema.StringAttribute{
		MarkdownDescription: "Type of the managed database instance",
		Computed:            true,
	}
	s.Attributes["zone"] = schema.StringAttribute{
		MarkdownDescription: "Zone where the instance resides, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
		Required:            true,
	}
	s.Attributes["primary_database"] = schema.StringAttribute{
		MarkdownDescription: "Primary database name",
		Computed:            true,
	}
	s.Attributes["additional_disk_space_gib"] = schema.Int64Attribute{
		MarkdownDescription: additionalDiskDescription,
		Computed:            true,
		Optional:            true,
	}

	s.Blocks["network"] = schema.SetNestedBlock{
		MarkdownDescription: "Private networks attached to the managed database",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "The name of the network. Must be unique within the service.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 65),
						stringvalidator.RegexMatches(
							regexp.MustCompile("^[a-zA-Z0-9_-]+$"), "must only contain alphanumeric characters, underscores, and hyphens"),
					},
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "The type of the network. Must be private.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							string(upcloud.LoadBalancerNetworkTypePrivate),
						),
					},
				},
				"family": schema.StringAttribute{
					MarkdownDescription: "Network family. Currently only `IPv4` is supported.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							string(upcloud.LoadBalancerAddressFamilyIPv4),
						),
					},
				},
				"uuid": schema.StringAttribute{
					Description: "Private network UUID. Must reside in the same zone as the database.",
					Required:    true,
				},
			},
		},
	}
	s.Blocks["properties"] = properties.GetBlock(dbType)
}

func networksFromPlan(ctx context.Context, data *databaseCommonModel) ([]upcloud.ManagedDatabaseNetwork, diag.Diagnostics) {
	var respDiagnostics diag.Diagnostics

	var networks []databaseNetworkModel
	respDiagnostics.Append(data.Network.ElementsAs(ctx, &networks, false)...)

	req := make([]upcloud.ManagedDatabaseNetwork, 0)
	for _, network := range networks {
		uuid := network.UUID.ValueString()
		r := upcloud.ManagedDatabaseNetwork{
			Name:   network.Name.ValueString(),
			Type:   network.Type.ValueString(),
			Family: network.Family.ValueString(),
			UUID:   &uuid,
		}

		req = append(req, r)
	}

	return req, respDiagnostics
}
