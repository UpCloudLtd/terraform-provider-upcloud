package server

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/storage"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	serverDescription = `The UpCloud server resource allows the creation, update and deletion of [cloud servers](https://upcloud.com/products/cloud-servers).

-> To deploy a GPU server, select a plan with ` + "`" + `GPU-` + "`" + ` prefix, e.g., ` + "`" + `GPU-8xCPU-64GB-1xL40S` + "`" + `. Use ` + "`" + `upctl zone devices` + "`" + ` command to list per zone GPU availability.`
	serverTitleLength       int = 255
	simpleBackupDescription     = `Simple backup schedule configuration

    The simple backups provide a simplified way to back up *all* of the storages attached to a given server. This means you cannot have simple backup set for a server, and individual ` + "`" + `backup_rules` + "`" + ` on the storages attached to the server. Such configuration will throw an error during execution. This also applies to ` + "`" + `backup_rules` + "`" + ` defined for server templates.

    ` + storage.BackupRuleSimpleBackupWarning
)

var (
	_ resource.Resource                 = &serverResource{}
	_ resource.ResourceWithConfigure    = &serverResource{}
	_ resource.ResourceWithImportState  = &serverResource{}
	_ resource.ResourceWithModifyPlan   = &serverResource{}
	_ resource.ResourceWithUpgradeState = &serverResource{}
)

func NewServerResource() resource.Resource {
	return &serverResource{}
}

type serverResource struct {
	client *service.Service
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

// Configure adds the provider configured client to the resource.
func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type serverModel struct {
	ID                types.String `tfsdk:"id"`
	Hostname          types.String `tfsdk:"hostname"`
	Title             types.String `tfsdk:"title"`
	Zone              types.String `tfsdk:"zone"`
	ServerGroup       types.String `tfsdk:"server_group"`
	Firewall          types.Bool   `tfsdk:"firewall"`
	Metadata          types.Bool   `tfsdk:"metadata"`
	CPU               types.Int64  `tfsdk:"cpu"`
	Mem               types.Int64  `tfsdk:"mem"`
	Timezone          types.String `tfsdk:"timezone"`
	VideoModel        types.String `tfsdk:"video_model"`
	NICModel          types.String `tfsdk:"nic_model"`
	Tags              types.Set    `tfsdk:"tags"`
	Host              types.Int64  `tfsdk:"host"`
	NetworkInterfaces types.List   `tfsdk:"network_interface"`
	Labels            types.Map    `tfsdk:"labels"`
	UserData          types.String `tfsdk:"user_data"`
	Plan              types.String `tfsdk:"plan"`
	StorageDevices    types.Set    `tfsdk:"storage_devices"`
	Template          types.List   `tfsdk:"template"`
	Login             types.Set    `tfsdk:"login"`
	SimpleBackup      types.Set    `tfsdk:"simple_backup"`
	BootOrder         types.String `tfsdk:"boot_order"`
	HotResize         types.Bool   `tfsdk:"hot_resize"`
}

type networkInterfaceModel struct {
	Index                 types.Int64  `tfsdk:"index"`
	IPAddressFamily       types.String `tfsdk:"ip_address_family"`
	IPAddress             types.String `tfsdk:"ip_address"`
	IPAddressFloating     types.Bool   `tfsdk:"ip_address_floating"`
	AdditionalIPAddresses types.Set    `tfsdk:"additional_ip_address"`
	MACAddress            types.String `tfsdk:"mac_address"`
	Type                  types.String `tfsdk:"type"`
	Network               types.String `tfsdk:"network"`
	SourceIPFiltering     types.Bool   `tfsdk:"source_ip_filtering"`
	Bootable              types.Bool   `tfsdk:"bootable"`
}

type additionalIPAddressModel struct {
	IPAddressFamily   types.String `tfsdk:"ip_address_family"`
	IPAddress         types.String `tfsdk:"ip_address"`
	IPAddressFloating types.Bool   `tfsdk:"ip_address_floating"`
}

func (m additionalIPAddressModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address_family":   types.StringType,
		"ip_address":          types.StringType,
		"ip_address_floating": types.BoolType,
	}
}

type storageDeviceModel struct {
	Address         types.String `tfsdk:"address"`
	AddressPosition types.String `tfsdk:"address_position"`
	Storage         types.String `tfsdk:"storage"`
	Type            types.String `tfsdk:"type"`
}

type templateModel struct {
	ID                     types.String `tfsdk:"id"`
	Address                types.String `tfsdk:"address"`
	AddressPosition        types.String `tfsdk:"address_position"`
	Encrypt                types.Bool   `tfsdk:"encrypt"`
	Size                   types.Int64  `tfsdk:"size"`
	Tier                   types.String `tfsdk:"tier"`
	Title                  types.String `tfsdk:"title"`
	Storage                types.String `tfsdk:"storage"`
	BackupRule             types.List   `tfsdk:"backup_rule"`
	FilesystemAutoresize   types.Bool   `tfsdk:"filesystem_autoresize"`
	DeleteAutoresizeBackup types.Bool   `tfsdk:"delete_autoresize_backup"`
}

type loginModel struct {
	User             types.String `tfsdk:"user"`
	Keys             types.List   `tfsdk:"keys"`
	CreatePassword   types.Bool   `tfsdk:"create_password"`
	PasswordDelivery types.String `tfsdk:"password_delivery"`
}

type simpleBackupModel struct {
	Plan types.String `tfsdk:"plan"`
	Time types.String `tfsdk:"time"`
}

func (r *serverResource) getSchema(version int64) schema.Schema {
	return schema.Schema{
		Version:             version,
		MarkdownDescription: serverDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname of the server.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					validatorutil.IsDomainName(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A short, informational description of the server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, serverTitleLength),
				},
			},
			"zone": schema.StringAttribute{
				Description: "The zone in which the server will be hosted, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_group": schema.StringAttribute{
				Description: "The UUID of a server group to attach this server to. Note that the server can also be attached to a server group via the `members` property of `upcloud_server_group`. Only one of the these should be defined at a time. This value is only updated if it has been set to non-zero value.",
				Optional:    true,
			},
			"firewall": schema.BoolAttribute{
				Description: "Are firewall rules active for the server",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.BoolAttribute{
				Description: "Is metadata service active for the server. The metadata service must be enabled when using recent cloud-init based templates.",
				Optional:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: "The number of CPU cores for the server",
				Optional:    true,
				Computed:    true,
			},
			"mem": schema.Int64Attribute{
				Description: "The amount of memory for the server (in megabytes)",
				Optional:    true,
				Computed:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "The timezone of the server. The timezone must be a valid timezone string, e.g. `Europe/Helsinki`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"video_model": schema.StringAttribute{
				Description: "The model of the server's video interface",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						upcloud.VideoModelCirrus,
						upcloud.VideoModelVGA,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"nic_model": schema.StringAttribute{
				Description: "The model of the server's network interfaces",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						upcloud.NICModelE1000,
						upcloud.NICModelVirtio,
						upcloud.NICModelRTL8139,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				Description: "The server related tags",
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					noDuplicateTagsValidator{},
				},
			},
			"host": schema.Int64Attribute{
				Description: "Use this to start the VM on a specific host. Refers to value from host -attribute. Only available for private cloud hosts",
				Computed:    true,
				Optional:    true,
			},
			"labels": utils.LabelsAttribute("server"),
			"user_data": schema.StringAttribute{
				Description: "Defines URL for a server setup script, or the script body itself",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan": schema.StringAttribute{
				Description: "The pricing plan used for the server. You can list available server plans with `upctl server plans`",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("cpu"),
						path.MatchRoot("mem"),
					),
				},
			},
			"boot_order": schema.StringAttribute{
				Description: "The boot device order, `cdrom`|`disk`|`network` or comma separated combination of those values. Defaults to `disk`",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hot_resize": schema.BoolAttribute{
				Description: "If set to true, allows changing the server plan without requiring a reboot. This enables hot resizing of the server. If hot resizing fails, the apply operation will fail.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"network_interface": schema.ListNestedBlock{
				Description: `One or more blocks describing the network interfaces of the server.

    In addition to list order, the configured network interfaces are matched to the server's actual network interfaces by ` + "`" + `index` + "`" + ` and ` + "`" + `ip_address` + "`" + ` fields. This is to avoid public and utility network interfaces being re-assigned when the server is updated. This might result to inaccurate diffs in the plan, when interfaces are re-ordered or when interface is removed from the middle of the list.

    We recommend explicitly setting the value for ` + "`" + `index` + "`" + ` in configuration, when re-ordering interfaces or when removing interface from middle of the list.`,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Description: "The interface index.",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"ip_address_family":   attributeIPAddressFamily("The type of the primary IP address of this interface (one of `IPv4` or `IPv6`)."),
						"ip_address":          attributeIPAddress("The primary IP address of this interface."),
						"ip_address_floating": attributeIPAddressFloating("`true` indicates that the primary IP address is a floating IP address."),
						"mac_address": schema.StringAttribute{
							Description: "The MAC address of the interface.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Network interface type. For private network interfaces, a network must be specified with an existing network id.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									upcloud.NetworkTypePrivate,
									upcloud.NetworkTypeUtility,
									upcloud.NetworkTypePublic,
								),
							},
						},
						"network": schema.StringAttribute{
							Description: "The UUID of the network to attach this interface to. Required for private network interfaces.",
							Optional:    true,
							Computed:    true,
						},
						"source_ip_filtering": schema.BoolAttribute{
							Description: "`true` if source IP should be filtered.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
						"bootable": schema.BoolAttribute{
							Description: "`true` if this interface should be used for network booting.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
					Blocks: map[string]schema.Block{
						"additional_ip_address": schema.SetNestedBlock{
							Description: "0-31 blocks of additional IP addresses to assign to this interface. Allowed only with network interfaces of type `private`",
							Validators: []validator.Set{
								setvalidator.SizeAtMost(31),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"ip_address_family":   attributeIPAddressFamily("The type of the additional IP address of this interface (one of `IPv4` or `IPv6`)."),
									"ip_address":          attributeIPAddress("An additional IP address for this interface."),
									"ip_address_floating": attributeIPAddressFloating("`true` indicates that the additional IP address is a floating IP address."),
								},
							},
						},
					},
				},
			},
			"storage_devices": schema.SetNestedBlock{
				Description: "A set of storage devices associated with the server",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "The device address the storage will be attached to (`scsi`|`virtio`|`ide`). Leave `address_position` field empty to auto-select next available address from that bus.",
							Computed:    true,
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("scsi", "virtio", "ide"),
							},
						},
						"address_position": schema.StringAttribute{
							Description: "The device position in the given bus (defined via field `address`). Valid values for address `virtio` are `0-15` (`0`, for example). Valid values for `scsi` or `ide` are `0-1:0-1` (`0:0`, for example). Leave empty to auto-select next available address in the given bus.",
							Computed:    true,
							Optional:    true,
						},
						"storage": schema.StringAttribute{
							Description: "The UUID of the storage to attach to the server.",
							Optional:    true,
						},
						"type": schema.StringAttribute{
							Description: "The device type the storage will be attached as",
							Computed:    true,
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("disk", "cdrom"),
							},
						},
					},
				},
			},
			"template": schema.ListNestedBlock{
				Description: "Block describing the preconfigured operating system",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.AtLeastOneOf(
						path.MatchRoot("storage_devices"),
					),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the storage",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"address": schema.StringAttribute{
							Description: "The device address the storage will be attached to (`scsi`|`virtio`|`ide`). Leave `address_position` field empty to auto-select next available address from that bus.",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"address_position": schema.StringAttribute{
							Description: "The device position in the given bus (defined via field `address`). For example `0:0`, or `0`. Leave empty to auto-select next available address in the given bus.",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"encrypt": schema.BoolAttribute{
							Description: "Sets if the storage is encrypted at rest",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
								boolplanmodifier.RequiresReplace(),
							},
						},
						"size": schema.Int64Attribute{
							Description: "The size of the storage in gigabytes",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(10, 2048),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"tier": schema.StringAttribute{
							Description: "The storage tier to use.",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
								stringplanmodifier.RequiresReplace(),
							},
						},
						"title": schema.StringAttribute{
							Description: "A short, informative description",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
							},
						},
						"storage": schema.StringAttribute{
							Description: "A valid storage UUID or template name. You can list available public templates with `upctl storage list --public --template` and available private templates with `upctl storage list --template`.",
							// TODO: validate isrequired
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"filesystem_autoresize": schema.BoolAttribute{
							Description: `If set to true, provider will attempt to resize partition and filesystem when the size of template storage changes.
							Please note that before the resize attempt is made, backup of the storage will be taken. If the resize attempt fails, the backup will be used
							to restore the storage and then deleted. If the resize attempt succeeds, backup will be kept (unless delete_autoresize_backup option is set to true).
							Taking and keeping backups incure costs.`,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"delete_autoresize_backup": schema.BoolAttribute{
							Description: "If set to true, the backup taken before the partition and filesystem resize attempt will be deleted immediately after success.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
					Blocks: map[string]schema.Block{
						"backup_rule": storage.BackupRuleBlock(),
					},
				},
			},
			"login": schema.SetNestedBlock{
				Description: "Configure access credentials to the server",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.StringAttribute{
							Description: "Username to be create to access the server",
							Optional:    true,
						},
						"keys": schema.ListAttribute{
							Description: "A list of ssh keys to access the server",
							ElementType: types.StringType,
							Optional:    true,
						},
						"create_password": schema.BoolAttribute{
							Description: "Indicates a password should be create to allow access",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"password_delivery": schema.StringAttribute{
							Description: "The delivery method for the server's root password (one of `none`, `email` or `sms`)",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("none"),
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf("none", "email", "sms"),
							},
						},
					},
				},
			},
			"simple_backup": schema.SetNestedBlock{
				Description: simpleBackupDescription,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"plan": schema.StringAttribute{
							Description: "Simple backup plan. Accepted values: daily, dailies, weeklies, monthlies.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("daily", "dailies", "weeklies", "monthlies"),
							},
						},
						"time": schema.StringAttribute{
							Description: "Time of the day at which backup will be taken. Should be provided in a hhmm format (e.g. 2230).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^\d{4}$`), "Time must be 4 digits in a hhmm format"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.getSchema(2)
}

func (r *serverResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := r.getSchema(0)
	schemaV1 := r.getSchema(1)
	return map[int64]resource.StateUpgrader{
		// Add index value to network interfaces.
		0: {
			PriorSchema: &schemaV0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var data serverModel
				resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

				if resp.Diagnostics.HasError() {
					return
				}

				uuid := data.ID.ValueString()
				networking, err := r.client.GetServerNetworks(ctx, &request.GetServerNetworksRequest{ServerUUID: uuid})
				if err != nil {
					resp.Diagnostics.AddError("Failed to get server networks", err.Error())
					return
				}

				if networking == nil {
					resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
				}

				var networkInterfaces []networkInterfaceModel
				resp.Diagnostics.Append(data.NetworkInterfaces.ElementsAs(ctx, &networkInterfaces, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if len(networking.Interfaces) != len(networkInterfaces) {
					resp.Diagnostics.AddError("State is not up-to-date", "Network interfaces have been modified outside of Terraform, unable to migrate the state. Correct the drift between state and the resource to continue")
					return
				}

				for i, iface := range networkInterfaces {
					index, diags := getIndexFromNetworking(networking, iface)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					networkInterfaces[i].Index = types.Int64Value(int64(index))
				}

				var diags diag.Diagnostics
				data.NetworkInterfaces, diags = types.ListValueFrom(ctx, data.NetworkInterfaces.ElementType(ctx), networkInterfaces)
				resp.Diagnostics.Append(diags...)
				resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
			},
		},
		// Replace empty login username with null value.
		1: {
			PriorSchema: &schemaV1,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var data serverModel
				resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

				if resp.Diagnostics.HasError() {
					return
				}

				if !data.Login.IsNull() {
					var login []loginModel
					resp.Diagnostics.Append(data.Login.ElementsAs(ctx, &login, false)...)

					if len(login) > 0 && login[0].User.ValueString() == "" {
						login[0].User = types.StringNull()

						data.Login, _ = types.SetValueFrom(ctx, data.Login.ElementType(ctx), login)
					}
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
			},
		},
	}
}

func getIndexFromNetworking(networking *upcloud.Networking, iface networkInterfaceModel) (index int, diags diag.Diagnostics) {
	for _, n := range networking.Interfaces {
		if n.Type == iface.Type.ValueString() && n.MAC == iface.MACAddress.ValueString() {
			index = n.Index
			return
		}
	}

	diags.AddError("Unable to find index", fmt.Sprintf("Unable to find index for interface %s", iface.MACAddress.ValueString()))
	return
}

func attributeIPAddressFamily(description string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(upcloud.IPAddressFamilyIPv4),
		Validators: []validator.String{
			stringvalidator.OneOf(
				upcloud.IPAddressFamilyIPv4,
				upcloud.IPAddressFamilyIPv6,
			),
		},
	}
}

func attributeIPAddress(description string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		Validators:          []validator.String{validatorutil.NewFrameworkStringValidator(validation.IsIPAddress)},
	}
}

func attributeIPAddressFloating(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: description,
		Computed:            true,
		// Force the value to be unknown if not set. For some reason, the value is null by default in additional_ip_address block, which causes data-consistency errors if user has defined additional IP addresses.
		Default: utils.StaticUnknown{},
	}
}

func (r *serverResource) updateCPUandMemPlan(state, plan *serverModel) {
	if state.Plan.Equal(plan.Plan) {
		if plan.CPU.IsUnknown() {
			plan.CPU = state.CPU
		}
		if plan.Mem.IsUnknown() {
			plan.Mem = state.Mem
		}
	}
}

func (r *serverResource) updateTagsPlan(server *upcloud.ServerDetails, plan *serverModel) {
	if plan.Tags.IsUnknown() && len(server.Tags) == 0 {
		plan.Tags = types.SetNull(types.StringType)
	}
}

func setAddressPlan(plan *networkInterfaceModel, ip upcloud.IPAddress) {
	plan.IPAddress = types.StringValue(ip.Address)
	plan.IPAddressFamily = types.StringValue(ip.Family)
	plan.IPAddressFloating = utils.AsBool(ip.Floating)
}

func (r *serverResource) updateNetworkInterfacesPlan(ctx context.Context, server *upcloud.ServerDetails, state, plan *serverModel, resp *resource.ModifyPlanResponse) {
	var networkInterfacesPlan []networkInterfaceModel
	resp.Diagnostics.Append(plan.NetworkInterfaces.ElementsAs(ctx, &networkInterfacesPlan, false)...)

	var networkInterfacesState []networkInterfaceModel
	resp.Diagnostics.Append(state.NetworkInterfaces.ElementsAs(ctx, &networkInterfacesState, false)...)

	networkInterfaceChanges := matchInterfacesToPlan(server.Networking.Interfaces, networkInterfacesState, networkInterfacesPlan)
	for i, planIface := range networkInterfacesPlan {
		change := networkInterfaceChanges[i]
		if canModifyInterface(change.state, change.plan, change.api) {
			if ip := findIPAddress(*change.api, planIface.IPAddress.ValueString()); ip != nil {
				// Handle cases where IP address is defined in the config.
				setAddressPlan(&planIface, *ip)
			} else if planIface.AdditionalIPAddresses.IsNull() && len(change.api.IPAddresses) > 0 {
				// Handle cases where IP address is defined by the system.
				setAddressPlan(&planIface, change.api.IPAddresses[0])
			}

			if !planIface.AdditionalIPAddresses.IsNull() {
				var additionalIPAddresses []additionalIPAddressModel
				resp.Diagnostics.Append(planIface.AdditionalIPAddresses.ElementsAs(ctx, &additionalIPAddresses, false)...)
				for i, additionalIPAddress := range additionalIPAddresses {
					if ip := findIPAddress(*change.api, additionalIPAddress.IPAddress.ValueString()); ip != nil {
						additionalIPAddress.IPAddress = types.StringValue(ip.Address)
						additionalIPAddress.IPAddressFamily = types.StringValue(ip.Family)
						additionalIPAddress.IPAddressFloating = utils.AsBool(ip.Floating)
					}
					additionalIPAddresses[i] = additionalIPAddress
				}

				var diags diag.Diagnostics
				planIface.AdditionalIPAddresses, diags = types.SetValueFrom(ctx, planIface.AdditionalIPAddresses.ElementType(ctx), additionalIPAddresses)
				resp.Diagnostics.Append(diags...)
			}
			planIface.MACAddress = types.StringValue(change.api.MAC)
			planIface.Network = types.StringValue(change.api.Network)
		}
		networkInterfacesPlan[i] = planIface
	}

	var diags diag.Diagnostics
	plan.NetworkInterfaces, diags = types.ListValueFrom(ctx, plan.NetworkInterfaces.ElementType(ctx), networkInterfacesPlan)
	resp.Diagnostics.Append(diags...)
}

func (r *serverResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if plan == nil {
		// Do not validate config for destroy plans.
		return
	}

	var config serverModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	resp.Diagnostics.Append(validatePlan(ctx, r.client, config.Plan)...)
	resp.Diagnostics.Append(validateZone(ctx, r.client, config.Zone)...)

	var state *serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if state == nil {
		return
	}

	resp.Diagnostics.Append(validateTagsChangeRequiresMainAccount(ctx, r.client, state.Tags, plan.Tags)...)

	r.updateCPUandMemPlan(state, plan)

	uuid := plan.ID.ValueString()
	server, err := r.client.GetServerDetails(ctx, &request.GetServerDetailsRequest{UUID: uuid})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read server details when modifying plan",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	r.updateTagsPlan(server, plan)
	r.updateNetworkInterfacesPlan(ctx, server, state, plan, resp)

	// Host might change if server is migrated to different host, so update host here instead of using state for unknown.
	if !changeRequiresServerStop(*state, *plan) {
		plan.Host = types.Int64Value(int64(server.Host))
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func setValues(ctx context.Context, data *serverModel, server *upcloud.ServerDetails) diag.Diagnostics {
	var respDiagnostics, diags diag.Diagnostics

	data.ID = types.StringValue(server.UUID)
	data.Host = types.Int64Value(int64(server.Host))
	data.Hostname = types.StringValue(server.Hostname)
	data.Title = types.StringValue(server.Title)
	data.Zone = types.StringValue(server.Zone)
	data.CPU = types.Int64Value(int64(server.CoreNumber))
	data.Mem = types.Int64Value(int64(server.MemoryAmount))

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(server.Labels))
	respDiagnostics.Append(diags...)

	data.NICModel = types.StringValue(server.NICModel)
	data.Timezone = types.StringValue(server.Timezone)
	data.VideoModel = types.StringValue(server.VideoModel)
	if !data.Metadata.IsNull() {
		data.Metadata = types.BoolValue(server.Metadata.Bool())
	}
	data.Plan = types.StringValue(server.Plan)
	data.BootOrder = types.StringValue(server.BootOrder)

	// Set hot_resize to false by default if not set
	if data.HotResize.IsNull() {
		data.HotResize = types.BoolValue(false)
	}

	if !data.Tags.IsNull() {
		data.Tags, diags = types.SetValueFrom(ctx, data.Tags.ElementType(ctx), server.Tags)
		respDiagnostics.Append(diags...)
	} else {
		data.Tags = types.SetNull(data.Tags.ElementType(ctx))
	}

	// Only track server_group when it has been configured to avoid changes when server is attached to group via upcloud_server_group.members.
	if !data.ServerGroup.IsNull() {
		data.ServerGroup = types.StringValue(server.ServerGroup)
	}

	if server.Firewall == "on" {
		data.Firewall = types.BoolValue(true)
	} else {
		data.Firewall = types.BoolValue(false)
	}

	if server.SimpleBackup != "no" {
		parts := strings.Split(server.SimpleBackup, ",")
		if n := len(parts); n != 2 {
			diags.AddError("Invalid simple backup configuration", fmt.Sprintf("Expected 2 parts, got %d", n))
		}
		simpleBackup := []simpleBackupModel{
			{
				Time: types.StringValue(parts[0]),
				Plan: types.StringValue(parts[1]),
			},
		}
		data.SimpleBackup, diags = types.SetValueFrom(ctx, data.SimpleBackup.ElementType(ctx), simpleBackup)
		respDiagnostics.Append(diags...)
	}

	var networkInterfaces []networkInterfaceModel

	// Handle current network interfaces
	var prevNetworkInterfaces []networkInterfaceModel
	handledInterfaces := make(map[int]bool)

	respDiagnostics.Append(data.NetworkInterfaces.ElementsAs(ctx, &prevNetworkInterfaces, false)...)

	for i, nic := range prevNetworkInterfaces {
		var iface *upcloud.ServerInterface
		index := nic.Index.ValueInt64()
		if index == 0 && i < len(server.Networking.Interfaces) {
			iface = &server.Networking.Interfaces[i]
			index = int64(iface.Index)
		} else {
			iface = findInterface(server.Networking.Interfaces, int(index))
		}
		if iface == nil {
			continue
		}

		ni, diags := setInterfaceValues(ctx, (*upcloud.Interface)(iface), nic.IPAddress)
		respDiagnostics.Append(diags...)

		networkInterfaces = append(networkInterfaces, ni)
		handledInterfaces[int(index)] = true
	}

	// Handle new network interfaces. This is needed for importing state.
	for _, iface := range server.Networking.Interfaces {
		if handledInterfaces[iface.Index] {
			continue
		}

		ni, diags := setInterfaceValues(ctx, (*upcloud.Interface)(&iface), types.StringNull())
		respDiagnostics.Append(diags...)
		networkInterfaces = append(networkInterfaces, ni)
	}

	data.NetworkInterfaces, diags = types.ListValueFrom(ctx, data.NetworkInterfaces.ElementType(ctx), networkInterfaces)
	respDiagnostics.Append(diags...)

	var storageDevices []storageDeviceModel
	template, diags := getTemplate(ctx, *data)
	respDiagnostics.Append(diags...)

	for _, serverStorage := range server.StorageDevices {
		if template != nil && serverStorage.UUID == template.ID.ValueString() {
			templates := []templateModel{
				{
					Address:         types.StringValue(utils.StorageAddressFormat(serverStorage.Address)),
					AddressPosition: types.StringValue(utils.StorageAddressPositionFormat(serverStorage.Address)),
					ID:              types.StringValue(serverStorage.UUID),
					Encrypt:         types.BoolValue(serverStorage.Encrypted.Bool()),
					Size:            types.Int64Value(int64(serverStorage.Size)),
					Tier:            types.StringValue(serverStorage.Tier),
					Title:           types.StringValue(serverStorage.Title),

					// Passthrough values
					Storage:                template.Storage,
					BackupRule:             template.BackupRule,
					FilesystemAutoresize:   template.FilesystemAutoresize,
					DeleteAutoresizeBackup: template.DeleteAutoresizeBackup,
				},
			}
			data.Template, diags = types.ListValueFrom(ctx, data.Template.ElementType(ctx), templates)
			respDiagnostics.Append(diags...)
		} else {
			storageDevices = append(storageDevices, storageDeviceModel{
				Address:         types.StringValue(utils.StorageAddressFormat(serverStorage.Address)),
				AddressPosition: types.StringValue(utils.StorageAddressPositionFormat(serverStorage.Address)),
				Storage:         types.StringValue(serverStorage.UUID),
				Type:            types.StringValue(serverStorage.Type),
			})
		}
	}

	data.StorageDevices, diags = types.SetValueFrom(ctx, data.StorageDevices.ElementType(ctx), storageDevices)
	respDiagnostics.Append(diags...)

	return respDiagnostics
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}
	var labelsSlice upcloud.LabelSlice = utils.LabelsMapToSlice(labels)

	title := data.Title.ValueString()
	if title == "" {
		title = defaultTitleFromHostname(data.Hostname.ValueString())
	}

	apiReq := &request.CreateServerRequest{
		BootOrder:    data.BootOrder.ValueString(),
		CoreNumber:   int(data.CPU.ValueInt64()),
		Host:         int(data.Host.ValueInt64()),
		Hostname:     data.Hostname.ValueString(),
		Labels:       &labelsSlice,
		MemoryAmount: int(data.Mem.ValueInt64()),
		Metadata:     utils.AsUpCloudBoolean(data.Metadata),
		NICModel:     data.NICModel.ValueString(),
		Plan:         data.Plan.ValueString(),
		ServerGroup:  data.ServerGroup.ValueString(),
		TimeZone:     data.Timezone.ValueString(),
		Title:        title,
		UserData:     data.UserData.ValueString(),
		VideoModel:   data.VideoModel.ValueString(),
		Zone:         data.Zone.ValueString(),
	}

	if !data.Firewall.IsNull() {
		if data.Firewall.ValueBool() {
			apiReq.Firewall = "on"
		} else {
			apiReq.Firewall = "off"
		}
	}

	if !data.Login.IsNull() {
		var login []loginModel
		resp.Diagnostics.Append(data.Login.ElementsAs(ctx, &login, false)...)

		loginOpts, deliveryMethod, diags := buildLoginOpts(ctx, login[0])
		resp.Diagnostics.Append(diags...)

		apiReq.LoginUser = loginOpts
		apiReq.PasswordDelivery = deliveryMethod
	}

	if !data.SimpleBackup.IsNull() {
		simpleBackup, diags := getSimpleBackup(ctx, data)
		resp.Diagnostics.Append(diags...)

		apiReq.SimpleBackup = buildSimpleBackupOpts(simpleBackup)
	}

	template, diags := getTemplate(ctx, data)
	resp.Diagnostics.Append(diags...)
	if template != nil {
		title := template.Title.ValueString()
		if title == "" {
			title = fmt.Sprintf("terraform-%s-disk", data.Hostname.ValueString())
		}

		storageDevice := request.CreateServerStorageDevice{
			Action: "clone",
			Address: buildStorageDeviceAddress(
				template.Address.ValueString(),
				template.AddressPosition.ValueString(),
			),
			Encrypted: upcloud.FromBool(template.Encrypt.ValueBool()),
			Size:      int(template.Size.ValueInt64()),
			Storage:   template.Storage.ValueString(),
			Tier:      template.Tier.ValueString(),
			Title:     title,
		}

		if !template.BackupRule.IsNull() {
			var backupRules []storage.BackupRuleModel
			resp.Diagnostics.Append(template.BackupRule.ElementsAs(ctx, &backupRules, false)...)

			storageDevice.BackupRule = storage.BackupRule(backupRules[0])
		}

		if source := template.Storage.ValueString(); source != "" {
			// Assume template name is given and attempt map name to UUID
			if _, err := uuid.ParseUUID(source); err != nil {
				l, err := r.client.GetStorages(ctx, &request.GetStoragesRequest{
					Type: upcloud.StorageTypeTemplate,
				})
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to get storages",
						utils.ErrorDiagnosticDetail(err),
					)
					return
				}
				for _, s := range l.Storages {
					if s.Title == source {
						source = s.UUID
						break
					}
				}
			}

			storageDevice.Storage = source
		}

		apiReq.StorageDevices = append(apiReq.StorageDevices, storageDevice)
	}

	if !data.StorageDevices.IsNull() {
		var storageDevices []storageDeviceModel
		resp.Diagnostics.Append(data.StorageDevices.ElementsAs(ctx, &storageDevices, false)...)

		for _, storageDevice := range storageDevices {
			apiReq.StorageDevices = append(apiReq.StorageDevices, request.CreateServerStorageDevice{
				Action: "attach",
				Address: buildStorageDeviceAddress(
					storageDevice.Address.ValueString(),
					storageDevice.AddressPosition.ValueString(),
				),
				Storage: storageDevice.Storage.ValueString(),
				Type:    storageDevice.Type.ValueString(),
			})
		}
	}

	networking, diags := buildNetworkOpts(ctx, data)
	resp.Diagnostics.Append(diags...)

	apiReq.Networking = &request.CreateServerNetworking{
		Interfaces: networking,
	}

	server, err := r.client.CreateServer(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create server",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	// Set template UUID as setValues uses that to determine if storage is a template.
	if template != nil {
		templates := []templateModel{
			*template,
		}
		templates[0].ID = types.StringValue(server.StorageDevices[0].UUID)
		data.Template, diags = types.ListValueFrom(ctx, data.Template.ElementType(ctx), templates)
		resp.Diagnostics.Append(diags...)
	}

	if !data.Tags.IsNull() {
		tags, diags := getTags(ctx, data.Tags)
		resp.Diagnostics.Append(diags...)

		if err := addTags(
			ctx,
			r.client,
			server.UUID,
			tags,
		); err != nil {
			resp.Diagnostics.AddError("Unable to add server tags", utils.ErrorDiagnosticDetail(err))
		}
	}

	server, err = r.client.WaitForServerState(ctx, &request.WaitForServerStateRequest{
		UUID:         server.UUID,
		DesiredState: upcloud.ServerStateStarted,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for server to be in started state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, server)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	server, err := r.client.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read server details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &data, server)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, state, plan serverModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	uuid := plan.ID.ValueString()
	serverDetails, err := r.client.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: uuid,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to get server details", utils.ErrorDiagnosticDetail(err))
		return
	}

	// Before stopping, validate network interface requests to avoid unnecessary server downtime
	resp.Diagnostics.Append(validateNetworkInterfaces(ctx, r.client, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine what changes are needed
	planChanged := !plan.Plan.Equal(state.Plan)
	hotResizeEnabled := plan.HotResize.ValueBool()
	needsServerStop := changeRequiresServerStop(state, plan)

	// Server stop is required for certain changes - this takes precedence over hot resize
	if needsServerStop {
		err := utils.VerifyServerStopped(ctx, request.StopServerRequest{
			UUID: uuid,
		}, r.client)
		if err != nil {
			resp.Diagnostics.AddError("Unable to stop server", utils.ErrorDiagnosticDetail(err))
			return
		}
	} else if planChanged && hotResizeEnabled {
		// Only attempt hot resize if no server stop is required
		hotResizeReq := &request.ModifyServerRequest{
			UUID: uuid,
		}

		// Set the appropriate fields based on whether it's a custom plan
		if plan.Plan.ValueString() == customPlanName {
			hotResizeReq.CoreNumber = int(plan.CPU.ValueInt64())
			hotResizeReq.MemoryAmount = int(plan.Mem.ValueInt64())
		} else {
			hotResizeReq.Plan = plan.Plan.ValueString()
		}

		_, err := r.client.ModifyServer(ctx, hotResizeReq)
		if err != nil {
			// If hot resize fails, return an error
			resp.Diagnostics.AddError(
				"Hot resize failed",
				fmt.Sprintf("Unable to hot resize server: %s. Hot resize requires the server to be in a state where it can be resized without a reboot. If hot resize is not possible, set hot_resize = false to use the standard approach with server reboot.", utils.ErrorDiagnosticDetail(err)),
			)
			return
		}
	}

	firewall := ""
	if !plan.Firewall.IsNull() && !plan.Firewall.IsUnknown() {
		if plan.Firewall.ValueBool() {
			firewall = "on"
		} else {
			firewall = "off"
		}
	}

	var labels map[string]string
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	}
	var labelsSlice upcloud.LabelSlice = utils.LabelsMapToSlice(labels)

	apiReq := &request.ModifyServerRequest{
		UUID: uuid,

		CoreNumber:   int(plan.CPU.ValueInt64()),
		Firewall:     firewall,
		Hostname:     plan.Hostname.ValueString(),
		Labels:       &labelsSlice,
		MemoryAmount: int(plan.Mem.ValueInt64()),
		Metadata:     utils.AsUpCloudBoolean(plan.Metadata),
		NICModel:     plan.NICModel.ValueString(),
		Plan:         plan.Plan.ValueString(),
		TimeZone:     plan.Timezone.ValueString(),
		Title:        plan.Title.ValueString(),
		VideoModel:   plan.VideoModel.ValueString(),
	}

	templateState, diags := getTemplate(ctx, state)
	resp.Diagnostics.Append(diags...)

	templatePlan, diags := getTemplate(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if !plan.SimpleBackup.Equal(state.SimpleBackup) && templatePlan != nil {
		replaced, diags := hasTemplateBackupRuleBeenReplacedWithSimpleBackups(ctx, state, plan)
		resp.Diagnostics.Append(diags...)
		if replaced {
			template, err := r.client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{UUID: templatePlan.ID.ValueString()})
			if err != nil {
				resp.Diagnostics.AddError("Unable to get template storage details", utils.ErrorDiagnosticDetail(err))
				return
			}

			if template.BackupRule != nil && template.BackupRule.Interval != "" {
				if _, err := r.client.ModifyStorage(ctx, &request.ModifyStorageRequest{
					UUID:       template.UUID,
					BackupRule: &upcloud.BackupRule{},
				}); err != nil {
					resp.Diagnostics.AddError("Unable to remove backup rule from template storage", utils.ErrorDiagnosticDetail(err))
					return
				}
			}
		}

		simpleBackup, diags := getSimpleBackup(ctx, plan)
		resp.Diagnostics.Append(diags...)
		apiReq.SimpleBackup = buildSimpleBackupOpts(simpleBackup)
	}

	if _, err := r.client.ModifyServer(ctx, apiReq); err != nil {
		resp.Diagnostics.AddError("Unable to modify server", utils.ErrorDiagnosticDetail(err))
		return
	}

	if !plan.ServerGroup.Equal(state.ServerGroup) {
		err = removeServerFromGroup(ctx, r.client, uuid, state.ServerGroup.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to remove server from server group", utils.ErrorDiagnosticDetail(err))
		}

		err = addServerToGroup(ctx, r.client, uuid, plan.ServerGroup.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to add server to server group", utils.ErrorDiagnosticDetail(err))
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		oldTags, diags := getTags(ctx, state.Tags)
		resp.Diagnostics.Append(diags...)

		newTags, diags := getTags(ctx, plan.Tags)
		resp.Diagnostics.Append(diags...)

		if err := updateTags(
			ctx,
			r.client,
			uuid,
			oldTags,
			newTags,
		); err != nil {
			if errors.As(err, &tagsExistsWarning{}) {
				resp.Diagnostics.AddWarning("Unable to create tags with matching letter casing", err.Error())
			} else {
				resp.Diagnostics.AddError("Unable to update server tags", utils.ErrorDiagnosticDetail(err))
			}
		}
	}

	if templateState != nil && templatePlan != nil {
		if !templatePlan.Title.Equal(templateState.Title) ||
			!templatePlan.Size.Equal(templateState.Size) ||
			!templatePlan.BackupRule.Equal(templateState.BackupRule) {
			apiReq := &request.ModifyStorageRequest{
				UUID:  templatePlan.ID.ValueString(),
				Size:  int(templatePlan.Size.ValueInt64()),
				Title: templatePlan.Title.ValueString(),
			}

			replaced, diags := hasTemplateBackupRuleBeenReplacedWithSimpleBackups(ctx, state, plan)
			resp.Diagnostics.Append(diags...)

			if !templatePlan.BackupRule.Equal(templateState.BackupRule) && !replaced {
				var backupRules []storage.BackupRuleModel
				resp.Diagnostics.Append(templatePlan.BackupRule.ElementsAs(ctx, &backupRules, false)...)

				apiReq.BackupRule = storage.BackupRule(backupRules[0])
			}

			storageDetails, err := r.client.ModifyStorage(ctx, apiReq)
			if err != nil {
				resp.Diagnostics.AddError("Unable to modify template storage", utils.ErrorDiagnosticDetail(err))
			}

			if !templatePlan.Size.Equal(templateState.Size) && templatePlan.FilesystemAutoresize.ValueBool() {
				resp.Diagnostics.Append(storage.ResizeStoragePartitionAndFs(
					ctx,
					r.client,
					storageDetails.UUID,
					templatePlan.DeleteAutoresizeBackup.ValueBool(),
				)...)
			}
		}

		if !templatePlan.Address.Equal(templateState.Address) || !templatePlan.AddressPosition.Equal(templateState.AddressPosition) {
			oldAddress := buildStorageDeviceAddress(
				templateState.Address.ValueString(),
				templateState.AddressPosition.ValueString(),
			)
			newAddress := buildStorageDeviceAddress(
				templatePlan.Address.ValueString(),
				templatePlan.AddressPosition.ValueString(),
			)

			if _, err := r.client.DetachStorage(ctx, &request.DetachStorageRequest{
				ServerUUID: uuid,
				Address:    oldAddress,
			}); err != nil {
				resp.Diagnostics.AddError("Unable to detach storage", utils.ErrorDiagnosticDetail(err))
				return
			}

			if _, err := r.client.AttachStorage(ctx, &request.AttachStorageRequest{
				Address:     newAddress,
				ServerUUID:  uuid,
				StorageUUID: templatePlan.ID.ValueString(),
			}); err != nil {
				resp.Diagnostics.AddError("Unable to attach storage", utils.ErrorDiagnosticDetail(err))
				return
			}
		}
	}

	if !plan.StorageDevices.Equal(state.StorageDevices) {
		var oldStorageDevices, newStorageDevices []storageDeviceModel
		resp.Diagnostics.Append(state.StorageDevices.ElementsAs(ctx, &oldStorageDevices, false)...)
		resp.Diagnostics.Append(plan.StorageDevices.ElementsAs(ctx, &newStorageDevices, false)...)

		oldSet := newStorageDeviceSet(oldStorageDevices)
		newSet := newStorageDeviceSet(newStorageDevices)

		for _, storageDevice := range oldStorageDevices {
			serverStorageDevice := serverDetails.StorageDevice(storageDevice.Storage.ValueString())
			if newSet.includes(storageDevice) || serverStorageDevice == nil {
				continue
			}

			if _, err := r.client.DetachStorage(ctx, &request.DetachStorageRequest{
				ServerUUID: uuid,
				Address:    serverStorageDevice.Address,
			}); err != nil {
				resp.Diagnostics.AddError("Unable to detach storage", utils.ErrorDiagnosticDetail(err))
				return
			}

			// Remove backup rule from the detached storage, if it was a result of simple backup setting
			if !plan.SimpleBackup.IsNull() {
				if _, err := r.client.ModifyStorage(ctx, &request.ModifyStorageRequest{
					UUID:       serverStorageDevice.UUID,
					BackupRule: &upcloud.BackupRule{},
				}); err != nil {
					resp.Diagnostics.AddError("Unable to remove backup rule from storage", utils.ErrorDiagnosticDetail(err))
					return
				}
			}
		}

		for _, storageDevice := range newStorageDevices {
			if oldSet.includes(storageDevice) {
				continue
			}

			if _, err := r.client.AttachStorage(ctx, &request.AttachStorageRequest{
				ServerUUID: uuid,
				Address: buildStorageDeviceAddress(
					storageDevice.Address.ValueString(),
					storageDevice.AddressPosition.ValueString(),
				),
				StorageUUID: storageDevice.Storage.ValueString(),
				Type:        storageDevice.Type.ValueString(),
			}); err != nil {
				resp.Diagnostics.AddError("Unable to attach storage", utils.ErrorDiagnosticDetail(err))
				return
			}
		}
	}

	if !plan.NetworkInterfaces.Equal(state.NetworkInterfaces) {
		resp.Diagnostics.Append(updateServerNetworkInterfaces(ctx, r.client, &state, &plan)...)
		if diags.HasError() {
			return
		}
	}

	server, err := utils.VerifyServerStarted(ctx, request.StartServerRequest{UUID: uuid, Host: int(config.Host.ValueInt64())}, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to start server", utils.ErrorDiagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(setValues(ctx, &plan, server)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func getTemplate(ctx context.Context, data serverModel) (*templateModel, diag.Diagnostics) {
	if data.Template.IsNull() {
		return nil, nil
	}

	var templates []templateModel
	diags := data.Template.ElementsAs(ctx, &templates, false)

	if diags.HasError() {
		return nil, diags
	}

	if len(templates) == 1 {
		return &templates[0], diags
	}
	return nil, diags
}

func getSimpleBackup(ctx context.Context, data serverModel) (*simpleBackupModel, diag.Diagnostics) {
	var simpleBackup []simpleBackupModel
	diags := data.SimpleBackup.ElementsAs(ctx, &simpleBackup, false)

	if diags.HasError() {
		return nil, diags
	}

	if len(simpleBackup) == 1 {
		return &simpleBackup[0], diags
	}
	return nil, diags
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	uuid := data.ID.ValueString()
	if err := utils.VerifyServerStopped(ctx, request.StopServerRequest{UUID: uuid}, r.client); err != nil {
		resp.Diagnostics.AddError("Unable to stop server", utils.ErrorDiagnosticDetail(err))
	}

	var tags []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)

	// Delete tags that are not used by any other servers
	if err := removeTags(ctx, r.client, uuid, tags); err != nil {
		resp.Diagnostics.AddWarning("Unable to delete tags that will be unused after server deletion", utils.ErrorDiagnosticDetail(err))
	}

	// Delete server
	deleteServerRequest := &request.DeleteServerRequest{
		UUID: uuid,
	}
	if err := r.client.DeleteServer(ctx, deleteServerRequest); err != nil {
		resp.Diagnostics.AddError("Unable to delete server", utils.ErrorDiagnosticDetail(err))
	}

	template, diags := getTemplate(ctx, data)
	resp.Diagnostics.Append(diags...)
	// Delete server root disk
	if template != nil {
		deleteStorageRequest := &request.DeleteStorageRequest{
			UUID: template.ID.ValueString(),
		}
		if err := r.client.DeleteStorage(ctx, deleteStorageRequest); err != nil {
			resp.Diagnostics.AddError("Unable to delete server root disk", utils.ErrorDiagnosticDetail(err))
		}
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
