package server

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/service/storage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	serverTitleLength       int = 255
	simpleBackupDescription     = `Simple backup schedule configuration

    The simple backups provide a simplified way to back up *all* of the storages attached to a given server. This means you cannot have simple backup set for a server, and individual ` + "`" + `backup_rules` + "`" + ` on the storages attached to the server. Such configuration will throw an error during execution. This also applies to ` + "`" + `backup_rules` + "`" + ` defined for server templates.

    ` + storage.BackupRuleSimpleBackupWarning
)

func ResourceServer() *schema.Resource {
	s := resourceServerV0()
	s.SchemaVersion = 1
	s.StateUpgraders = []schema.StateUpgrader{
		{
			Type:    resourceServerV0().CoreConfigSchema().ImpliedType(),
			Upgrade: resourceServerStateUpgradeV0,
			Version: 0,
		},
	}

	return s
}

func resourceServerV0() *schema.Resource {
	return &schema.Resource{
		Description:   "The UpCloud server resource allows the creation, update and deletion of a [cloud server](https://upcloud.com/products/cloud-servers).",
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description: "A valid domain name",
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: validation.AllDiag(
					validation.ToDiagFunc(
						validation.StringLenBetween(1, 128),
					),
					validator.ValidateDomainNameDiag,
				),
			},
			"title": {
				Description:  "A short, informational description",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, serverTitleLength),
			},
			"zone": {
				Description: "The zone in which the server will be hosted, e.g. `de-fra1`. You can list available zones with `upctl zone list`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"server_group": {
				Type:        schema.TypeString,
				Description: "The UUID of a server group to attach this server to. Note that the server can also be attached to a server group via the `members` property of `upcloud_server_group`. Only one of the these should be defined at a time. This value is only updated if it has been set to non-zero value.",
				Default:     "",
				Optional:    true,
			},
			"firewall": {
				Description: "Are firewall rules active for the server",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"metadata": {
				Description: "Is the metadata service active for the server",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"cpu": {
				Description:   "The number of CPU for the server",
				Type:          schema.TypeInt,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"plan"},
			},
			"mem": {
				Description:   "The size of memory for the server (in megabytes)",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"plan"},
			},
			"timezone": {
				Description: "A timezone identifier, e.g. `Europe/Helsinki`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"video_model": {
				Description: "The model of the server's video interface",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateDiagFunc: func(v interface{}, _ cty.Path) diag.Diagnostics {
					switch v.(string) {
					case upcloud.VideoModelCirrus, upcloud.VideoModelVGA:
						return nil
					default:
						return diag.Diagnostics{diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "'video_model' has incorrect value",
							Detail: fmt.Sprintf(
								"'video_model' must be one of %s or %s",
								upcloud.VideoModelCirrus,
								upcloud.VideoModelVGA),
						}}
					}
				},
			},
			"nic_model": {
				Description: "The model of the server's network interfaces",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateDiagFunc: func(v interface{}, _ cty.Path) diag.Diagnostics {
					switch v.(string) {
					case upcloud.NICModelE1000, upcloud.NICModelVirtio, upcloud.NICModelRTL8139:
						return nil
					default:
						return diag.Diagnostics{diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "'nic_model' has incorrect value",
							Detail: fmt.Sprintf(
								"'nic_model' must be one of %s, %s or %s",
								upcloud.NICModelE1000,
								upcloud.NICModelVirtio,
								upcloud.NICModelRTL8139),
						}}
					}
				},
			},
			"tags": {
				Description: "The server related tags",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"host": {
				Description: "Use this to start the VM on a specific host. Refers to value from host -attribute. Only available for private cloud hosts",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"network_interface": {
				Type: schema.TypeList,
				Description: `One or more blocks describing the network interfaces of the server.

    In addition to list order, the configured network interfaces are matched to the server's actual network interfaces by ` + "`" + `index` + "`" + ` and ` + "`" + `ip_address` + "`" + ` fields. This is to avoid public and utility network interfaces being re-assigned when the server is updated. This might result to inaccurate diffs in the plan, when interfaces are re-ordered or when interface is removed from the middle of the list.

    We recommend explicitly setting the value for ` + "`" + `index` + "`" + ` in configuration, when re-ordering interfaces or when removing interface from middle of the list.`,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:         schema.TypeInt,
							Description:  "The interface index.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"ip_address_family":   schemaIPAddressFamily("The type of the primary IP address of this interface (one of `IPv4` or `IPv6`)."),
						"ip_address":          schemaIPAddress("The assigned primary IP address."),
						"ip_address_floating": schemaIPAddressFloating("`true` indicates that the primary IP address is a floating IP address."),
						"additional_ip_address": {
							Type:        schema.TypeSet,
							Description: "0-4 blocks of additional IP addresses to assign to this interface. Allowed only with network interfaces of type `private`",
							Optional:    true,
							MaxItems:    4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_address_family":   schemaIPAddressFamily("The type of this additional IP address of this interface (one of `IPv4` or `IPv6`)."),
									"ip_address":          schemaIPAddress("The assigned additional IP address."),
									"ip_address_floating": schemaIPAddressFloating("`true` indicates that the additional IP address is a floating IP address."),
								},
							},
						},
						"mac_address": {
							Type:        schema.TypeString,
							Description: "The assigned MAC address.",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Network interface type. For private network interfaces, a network must be specified with an existing network id.",
							Required:    true,
							ValidateDiagFunc: func(v interface{}, _ cty.Path) diag.Diagnostics {
								switch v.(string) {
								case upcloud.NetworkTypePrivate, upcloud.NetworkTypeUtility, upcloud.NetworkTypePublic:
									return nil
								default:
									return diag.Diagnostics{diag.Diagnostic{
										Severity: diag.Error,
										Summary:  "'type' has incorrect value",
										Detail: fmt.Sprintf(
											"'type' must be one of %s, %s or %s",
											upcloud.NetworkTypePrivate,
											upcloud.NetworkTypePublic,
											upcloud.NetworkTypeUtility),
									}}
								}
							},
						},
						"network": {
							Type:        schema.TypeString,
							Description: "The unique ID of a network to attach this network to.",
							Optional:    true,
							Computed:    true,
						},
						"source_ip_filtering": {
							Type:        schema.TypeBool,
							Description: "`true` if source IP should be filtered.",
							Optional:    true,
							Default:     true,
						},
						"bootable": {
							Type:        schema.TypeBool,
							Description: "`true` if this interface should be used for network booting.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"labels": utils.LabelsSchema("server"),
			"user_data": {
				Description: "Defines URL for a server setup script, or the script body itself",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"plan": {
				Description: "The pricing plan used for the server. You can list available server plans with `upctl server plans`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"storage_devices": {
				Description: "A list of storage devices associated with the server",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Description:  "The device address the storage will be attached to (`scsi`|`virtio`|`ide`). Leave `address_position` field empty to auto-select next available address from that bus.",
							Type:         schema.TypeString,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"scsi", "virtio", "ide"}, false),
						},
						"address_position": {
							Description: "The device position in the given bus (defined via field `address`). Valid values for address `virtio` are `0-15` (`0`, for example). Valid values for `scsi` or `ide` are `0-1:0-1` (`0:0`, for example). Leave empty to auto-select next available address in the given bus.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
						},
						"storage": {
							Description: "A valid storage UUID",
							Type:        schema.TypeString,
							Required:    true,
						},
						"type": {
							Description:  "The device type the storage will be attached as",
							Type:         schema.TypeString,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"disk", "cdrom"}, false),
						},
					},
				},
				Set: func(v interface{}) int {
					// compute a consistent hash for this TypeSet, mandatory
					m := v.(map[string]interface{})
					return schema.HashString(
						fmt.Sprintf("%s-%s-%s", m["storage"].(string), m["address"].(string), m["address_position"].(string)),
					)
				},
			},
			"template": {
				Description:  "Block describing the preconfigured operating system",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				AtLeastOneOf: []string{"storage_devices", "template"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The unique identifier for the storage",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"address": {
							Description: "The device address the storage will be attached to (`scsi`|`virtio`|`ide`). Leave `address_position` field empty to auto-select next available address from that bus.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
						},
						"address_position": {
							Description: "The device position in the given bus (defined via field `address`). For example `0:0`, or `0`. Leave empty to auto-select next available address in the given bus.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
						},
						"encrypt": {
							Description: "Sets if the storage is encrypted at rest",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
						"size": {
							Description:  "The size of the storage in gigabytes",
							Type:         schema.TypeInt,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.IntBetween(10, 2048),
						},
						// will be set to value matching the plan
						"tier": {
							Description: "The storage tier to use",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"title": {
							Description:  "A short, informative description",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"storage": {
							Description: "A valid storage UUID or template name. You can list available public templates with `upctl storage list --public --template` and available private templates with `upctl storage list --template`.",
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
						},
						"backup_rule": storage.BackupRuleSchema(),
						"filesystem_autoresize": {
							Description: `If set to true, provider will attempt to resize partition and filesystem when the size of template storage changes.
							Please note that before the resize attempt is made, backup of the storage will be taken. If the resize attempt fails, the backup will be used
							to restore the storage and then deleted. If the resize attempt succeeds, backup will be kept (unless delete_autoresize_backup option is set to true).
							Taking and keeping backups incure costs.`,
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"delete_autoresize_backup": {
							Description: "If set to true, the backup taken before the partition and filesystem resize attempt will be deleted immediately after success.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"login": {
				Description: "Configure access credentials to the server",
				Type:        schema.TypeSet,
				ForceNew:    true,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Description: "Username to be create to access the server",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"keys": {
							Description: "A list of ssh keys to access the server",
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"create_password": {
							Description: "Indicates a password should be create to allow access",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"password_delivery": {
							Description:  "The delivery method for the server's root password (one of `none`, `email` or `sms`)",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "none",
							ValidateFunc: validation.StringInSlice([]string{"none", "email", "sms"}, false),
						},
					},
				},
			},
			"simple_backup": {
				Description:   simpleBackupDescription,
				Type:          schema.TypeSet,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"template.0.backup_rule"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan": {
							Description:  "Simple backup plan. Accepted values: daily, dailies, weeklies, monthlies.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"daily", "dailies", "weeklies", "monthlies"}, false),
						},
						"time": {
							Description:  "Time of the day at which backup will be taken. Should be provided in a hhmm format (e.g. 2230).",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{4}$`), "Time must be 4 digits in a hhmm format"),
						},
					},
				},
			},
			"boot_order": {
				Description: "The boot device order, `cdrom`|`disk`|`network` or comma separated combination of those values. Defaults to `disk`",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
		},
		CustomizeDiff: customdiff.Sequence(
			// Validate tags here, because in-schema validation is only available for primitive types
			validateTagsChange,
		),
	}
}

func resourceServerStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	svc := meta.(*service.Service)
	networking, err := svc.GetServerNetworks(ctx, &request.GetServerNetworksRequest{ServerUUID: rawState["id"].(string)})
	if err != nil {
		return rawState, err
	}

	if networking == nil {
		return rawState, nil
	}

	rawInterfaces, ok := rawState["network_interface"].([]interface{})
	if !ok {
		return rawState, fmt.Errorf("network_interface not found in state")
	}

	if len(networking.Interfaces) != len(rawInterfaces) {
		return rawState, fmt.Errorf("values for network_interface have been modified outside of Terraform, unable to migrate the state. correct the drift between state and the resource to continue")
	}

	for k, v := range rawInterfaces {
		iface := v.(map[string]interface{})
		index, err := getIndexFromNetworking(networking, iface)
		if err != nil {
			return rawState, err
		}

		iface["index"] = index
		rawInterfaces[k] = iface
	}

	rawState["network_interface"] = rawInterfaces

	return rawState, nil
}

func getIndexFromNetworking(networking *upcloud.Networking, iface map[string]interface{}) (int, error) {
	for _, n := range networking.Interfaces {
		if n.Type == iface["type"].(string) && n.MAC == iface["mac_address"].(string) {
			return n.Index, nil
		}
	}

	return 0, fmt.Errorf("unable to find index for interface %s", iface["mac_address"].(string))
}

func schemaIPAddressFamily(description string) *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Description:      description,
		Optional:         true,
		Default:          upcloud.IPAddressFamilyIPv4,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{upcloud.IPAddressFamilyIPv4, upcloud.IPAddressFamilyIPv6}, false)),
	}
}

func schemaIPAddress(description string) *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Description:      description,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
	}
}

func schemaIPAddressFloating(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: description,
		Computed:    true,
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)
	diags := diag.Diagnostics{}

	if err := validatePlan(ctx, client, d.Get("plan").(string)); err != nil {
		return diag.FromErr(err)
	}

	if err := validateZone(ctx, client, d.Get("zone").(string)); err != nil {
		return diag.FromErr(err)
	}

	r, err := buildServerOpts(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("title"); ok {
		r.Title = d.Get("title").(string)
	} else {
		r.Title = defaultTitleFromHostname(d.Get("hostname").(string))
	}

	serverDetails, err := client.CreateServer(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serverDetails.UUID)
	tflog.Info(ctx, "server created", map[string]interface{}{"title": serverDetails.Title, "uuid": serverDetails.UUID})

	// set template id from the payload (if passed)
	if _, ok := d.GetOk("template.0"); ok {
		_ = d.Set("template", []map[string]interface{}{{
			"id":                       serverDetails.StorageDevices[0].UUID,
			"storage":                  d.Get("template.0.storage"),
			"filesystem_autoresize":    d.Get("template.0.filesystem_autoresize"),
			"delete_autoresize_backup": d.Get("template.0.delete_autoresize_backup"),
		}})
	}

	// Add server tags
	if tags, tagsExists := d.GetOk("tags"); tagsExists {
		tags := utils.ExpandStrings(tags)
		if err := addTags(ctx, client, serverDetails.UUID, tags); err != nil {
			if errors.As(err, &tagsExistsWarning{}) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  err.Error(),
				})
			} else {
				return diag.FromErr(err)
			}
		}
	}

	_, err = client.WaitForServerState(ctx, &request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return append(diags, resourceServerRead(ctx, d, meta)...)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)
	diags := diag.Diagnostics{}

	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(ctx, r)
	if err != nil {
		return utils.HandleResourceError(d.Get("hostname").(string), d, err)
	}
	_ = d.Set("hostname", server.Hostname)
	if server.Title != defaultTitleFromHostname(server.Hostname) {
		_ = d.Set("title", server.Title)
	} else {
		_ = d.Set("title", nil)
	}
	_ = d.Set("zone", server.Zone)
	_ = d.Set("cpu", server.CoreNumber)
	_ = d.Set("mem", server.MemoryAmount)

	_ = d.Set("labels", utils.LabelsSliceToMap(server.Labels))

	_ = d.Set("nic_model", server.NICModel)
	_ = d.Set("timezone", server.Timezone)
	_ = d.Set("video_model", server.VideoModel)
	_ = d.Set("metadata", server.Metadata.Bool())
	_ = d.Set("plan", server.Plan)
	_ = d.Set("boot_order", server.BootOrder)

	// XXX: server.Tags returns an empty slice rather than nil when it's empty
	if len(server.Tags) > 0 {
		_ = d.Set("tags", server.Tags)
	}

	// Only track server_group when it has been configured to avoid changes when server is attached to group via upcloud_server_group.members.
	if _, ok := d.GetOk("server_group"); ok {
		_ = d.Set("server_group", server.ServerGroup)
	}

	if server.Firewall == "on" {
		_ = d.Set("firewall", true)
	} else {
		_ = d.Set("firewall", false)
	}
	if server.SimpleBackup != "no" {
		p := strings.Split(server.SimpleBackup, ",")
		simpleBackup := map[string]interface{}{
			"time": p[0],
			"plan": p[1],
		}

		_ = d.Set("simple_backup", []interface{}{simpleBackup})
	}

	var connIP string
	n, ok := d.Get("network_interface.#").(int)
	if !ok {
		return diag.Errorf("unable to read network_interface count")
	}
	networkInterfaces := make([]map[string]interface{}, n)
	for i := 0; i < n; i++ {
		key := interfaceKey(i)
		val, ok := d.Get(key).(map[string]interface{})
		if !ok {
			return diag.Errorf("unable to read '%s' value", key)
		}

		var iface *upcloud.ServerInterface
		if index := val["index"].(int); index == 0 && i < len(server.Networking.Interfaces) {
			iface = &server.Networking.Interfaces[i]
		} else {
			iface = findInterface(server.Networking.Interfaces, val["index"].(int))
		}
		if iface == nil {
			continue
		}

		ni := setInterfaceValues((*upcloud.Interface)(iface), val["ip_address"])
		networkInterfaces[i] = ni

		if iface.Type == upcloud.NetworkTypePublic &&
			iface.IPAddresses[0].Family == upcloud.IPAddressFamilyIPv4 {
			connIP = iface.IPAddresses[0].Address
		}
	}

	if err := d.Set("network_interface", networkInterfaces); err != nil {
		return diag.FromErr(err)
	}

	storageDevices := []interface{}{}
	for _, serverStorage := range server.StorageDevices {
		// the template is managed within the server
		if serverStorage.UUID == d.Get("template.0.id") {
			_ = d.Set("template", []map[string]interface{}{{
				"address":          utils.StorageAddressFormat(serverStorage.Address),
				"address_position": utils.StorageAddressPositionFormat(serverStorage.Address),
				"id":               serverStorage.UUID,
				"encrypt":          serverStorage.Encrypted.Bool(),
				"size":             serverStorage.Size,
				"title":            serverStorage.Title,
				"storage":          d.Get("template.0.storage"),
				"tier":             serverStorage.Tier,
				// NOTE: backupRule cannot be derived from server.storageDevices payload, will not sync if changed elsewhere
				"backup_rule": d.Get("template.0.backup_rule"),
				// Those fields are not set anywhere in the API, they are just for internal TF use
				"filesystem_autoresize":    d.Get("template.0.filesystem_autoresize"),
				"delete_autoresize_backup": d.Get("template.0.delete_autoresize_backup"),
			}})
		} else {
			storageDevices = append(storageDevices, map[string]interface{}{
				"address":          utils.StorageAddressFormat(serverStorage.Address),
				"address_position": utils.StorageAddressPositionFormat(serverStorage.Address),
				"storage":          serverStorage.UUID,
				"type":             serverStorage.Type,
			})
		}
	}
	_ = d.Set("storage_devices", storageDevices)

	// Initialize the connection information.
	d.SetConnInfo(map[string]string{
		"host":     connIP,
		"password": "",
		"type":     "ssh",
		"user":     "root",
	})

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)
	diags := diag.Diagnostics{}

	planHasChange := d.HasChange("plan")
	if planHasChange {
		if err := validatePlan(ctx, client, d.Get("plan").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	serverDetails, err := client.GetServerDetails(ctx, &request.GetServerDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// Before stopping, validate network interface requests to avoid unnecessary server downtime
	err = validateNetworkInterfaces(ctx, client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Stop the server if the requested changes require it
	if d.HasChanges("cpu", "mem", "timezone", "nic_model", "video_model", "template.0.size", "storage_devices", "network_interface") || planHasChange {
		err := utils.VerifyServerStopped(ctx, request.StopServerRequest{
			UUID: d.Id(),
		},
			meta,
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	r := &request.ModifyServerRequest{
		UUID: d.Id(),
	}

	r.Hostname = d.Get("hostname").(string)
	if d.HasChange("labels") {
		r.Labels = buildLabels(d.Get("labels").(map[string]interface{}))
	}

	if attr, ok := d.GetOk("title"); ok {
		r.Title = attr.(string)
	} else {
		r.Title = defaultTitleFromHostname(d.Get("hostname").(string))
	}

	if attr, ok := d.GetOk("timezone"); ok {
		r.TimeZone = attr.(string)
	}

	if attr, ok := d.GetOk("video_model"); ok {
		r.VideoModel = attr.(string)
	}

	if attr, ok := d.GetOk("nic_model"); ok {
		r.NICModel = attr.(string)
	}

	r.Metadata = upcloud.FromBool(d.Get("metadata").(bool))

	if d.Get("firewall").(bool) {
		r.Firewall = "on"
	} else {
		r.Firewall = "off"
	}

	if d.HasChange("simple_backup") {
		if sb, ok := d.GetOk("simple_backup"); ok {
			// Special handling for a situation where user adds simple backup rule for the server
			// and removes backup_rule from a template with one apply. This needs to be done
			// to prevent backup rule conflict error. We do not need to check if user removed
			// template backup rule from the config, because having it together with server
			// simple backup is not allowed on schema level
			// Also see notes for simple_backup block in server resource docs for more insight:
			// https://github.com/UpCloudLtd/terraform-provider-upcloud/blob/master/docs/resources/server.md#nested-schema-for-simple_backup
			if hasTemplateBackupRuleBeenReplacedWithSimpleBackups(d) {
				templateID := d.Get("template.0.id").(string)

				tmpl, err := client.GetStorageDetails(ctx, &request.GetStorageDetailsRequest{UUID: templateID})
				if err != nil {
					return diag.FromErr(err)
				}

				if tmpl.BackupRule != nil && tmpl.BackupRule.Interval != "" {
					r := &request.ModifyStorageRequest{
						UUID:       templateID,
						BackupRule: &upcloud.BackupRule{},
					}

					if _, err := client.ModifyStorage(ctx, r); err != nil {
						return diag.FromErr(err)
					}
				}
			}

			simpleBackupAttrs := sb.(*schema.Set).List()[0].(map[string]interface{})
			r.SimpleBackup = buildSimpleBackupOpts(simpleBackupAttrs)
		} else {
			r.SimpleBackup = "no"
		}
	}

	if d.HasChanges("cpu", "mem") || planHasChange {
		if plan, ok := d.GetOk("plan"); ok && plan.(string) != "custom" {
			r.Plan = plan.(string)
		} else {
			r.CoreNumber = d.Get("cpu").(int)
			r.MemoryAmount = d.Get("mem").(int)
			r.Plan = "custom"
		}
	}

	if _, err := client.ModifyServer(ctx, r); err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("server_group") {
		oldGroup, newGroup := d.GetChange("server_group")

		err = removeServerFromGroup(ctx, client, d.Id(), oldGroup.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		err = addServerToGroup(ctx, client, d.Id(), newGroup.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		oldTags, newTags := d.GetChange("tags")
		if err := updateTags(
			ctx,
			client,
			d.Id(),
			utils.ExpandStrings(oldTags),
			utils.ExpandStrings(newTags),
		); err != nil {
			if errors.As(err, &tagsExistsWarning{}) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  err.Error(),
				})
			} else {
				return diag.FromErr(err)
			}
		}
	}

	// handle the template
	if d.HasChanges("template.0.title", "template.0.size", "template.0.backup_rule") {
		template := d.Get("template.0").(map[string]interface{})
		r := &request.ModifyStorageRequest{}

		r.UUID = template["id"].(string)
		r.Size = template["size"].(int)
		r.Title = template["title"].(string)

		if d.HasChange("template.0.backup_rule") && !hasTemplateBackupRuleBeenReplacedWithSimpleBackups(d) {
			if backupRule, ok := d.GetOk("template.0.backup_rule.0"); ok {
				rule := backupRule.(map[string]interface{})
				r.BackupRule = storage.BackupRule(rule)
			}
		}

		storageDetails, err := client.ModifyStorage(ctx, r)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("template.0.size") && d.Get("template.0.filesystem_autoresize").(bool) {
			diags = append(diags, utils.AsSDKv2Diags(storage.ResizeStoragePartitionAndFs(
				ctx,
				client,
				storageDetails.UUID,
				d.Get("template.0.delete_autoresize_backup").(bool),
			))...)
		}
	}

	// should reattach if address changed
	if d.HasChange("template.0.address") || d.HasChange("template.0.address_position") {
		oldAddress, newAddress := d.GetChange("template.0.address")
		oldPosition, newPosition := d.GetChange("template.0.address_position")

		detachAddress := utils.StorageAddressFormat(oldAddress.(string))
		if oldPosition.(string) != "" {
			detachAddress = fmt.Sprintf(":%s", oldPosition.(string))
		}

		if _, err := client.DetachStorage(ctx, &request.DetachStorageRequest{
			ServerUUID: d.Id(),
			Address:    detachAddress,
		}); err != nil {
			return diag.FromErr(err)
		}

		attachAddress := utils.StorageAddressFormat(newAddress.(string))
		if newPosition.(string) != "" {
			attachAddress = fmt.Sprintf(":%s", newPosition.(string))
		}

		if _, err := client.AttachStorage(ctx, &request.AttachStorageRequest{
			Address:     attachAddress,
			ServerUUID:  d.Id(),
			StorageUUID: d.Get("template.0.id").(string),
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	// handle the other storage devices
	if d.HasChange("storage_devices") {
		o, n := d.GetChange("storage_devices")

		// detach the devices that should be detached or should be re-attached with different parameters
		for _, rawStorageDevice := range o.(*schema.Set).Difference(n.(*schema.Set)).List() {
			storageDevice := rawStorageDevice.(map[string]interface{})
			serverStorageDevice := serverDetails.StorageDevice(storageDevice["storage"].(string))
			if serverStorageDevice == nil {
				continue
			}
			if _, err := client.DetachStorage(ctx, &request.DetachStorageRequest{
				ServerUUID: d.Id(),
				Address:    serverStorageDevice.Address,
			}); err != nil {
				return diag.FromErr(err)
			}

			// Remove backup rule from the detached storage, if it was a result of simple backup setting
			if _, ok := d.GetOk("simple_backup"); ok {
				if _, err := client.ModifyStorage(ctx, &request.ModifyStorageRequest{
					UUID:       serverStorageDevice.UUID,
					BackupRule: &upcloud.BackupRule{},
				}); err != nil {
					return diag.FromErr(err)
				}
			}
		}

		// attach the storages that are new or have changed
		for _, rawStorageDevice := range n.(*schema.Set).Difference(o.(*schema.Set)).List() {
			storageDevice := rawStorageDevice.(map[string]interface{})
			address := storageDevice["address"].(string)
			position := storageDevice["address_position"].(string)
			if position != "" {
				address += fmt.Sprintf(":%s", position)
			}
			if _, err := client.AttachStorage(ctx, &request.AttachStorageRequest{
				ServerUUID:  d.Id(),
				Address:     address,
				StorageUUID: storageDevice["storage"].(string),
				Type:        storageDevice["type"].(string),
			}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("network_interface") {
		if err := updateServerNetworkInterfaces(ctx, client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := utils.VerifyServerStarted(ctx, request.StartServerRequest{UUID: d.Id(), Host: d.Get("host").(int)}, meta); err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, resourceServerRead(ctx, d, meta)...)

	return diags
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	// Verify server is stopped before deletion
	if err := utils.VerifyServerStopped(ctx, request.StopServerRequest{UUID: d.Id()}, meta); err != nil {
		return diag.FromErr(err)
	}

	// Delete tags that are not used by any other servers
	if err := removeTags(ctx, client, d.Id(), utils.ExpandStrings(d.Get("tags"))); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "failed to delete tags that will be unused after server deletion",
			Detail:   err.Error(),
		})
	}

	// Delete server
	deleteServerRequest := &request.DeleteServerRequest{
		UUID: d.Id(),
	}
	tflog.Info(ctx, "deleting server", map[string]interface{}{"uuid": d.Id()})
	if err := client.DeleteServer(ctx, deleteServerRequest); err != nil {
		return diag.FromErr(err)
	}

	// Delete server root disk
	if template, ok := d.GetOk("template.0"); ok {
		template := template.(map[string]interface{})
		deleteStorageRequest := &request.DeleteStorageRequest{
			UUID: template["id"].(string),
		}
		tflog.Info(ctx, "deleting server storage", map[string]interface{}{"storage_uuid": deleteStorageRequest.UUID})
		if err := client.DeleteStorage(ctx, deleteStorageRequest); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func buildServerOpts(ctx context.Context, d *schema.ResourceData, meta interface{}) (*request.CreateServerRequest, error) {
	r := &request.CreateServerRequest{
		Zone:     d.Get("zone").(string),
		Hostname: d.Get("hostname").(string),
	}

	if attr, ok := d.GetOk("server_group"); ok {
		r.ServerGroup = attr.(string)
	}
	if attr, ok := d.GetOk("firewall"); ok {
		if attr.(bool) {
			r.Firewall = "on"
		} else {
			r.Firewall = "off"
		}
	}
	if attr, ok := d.GetOk("labels"); ok {
		r.Labels = buildLabels(attr.(map[string]interface{}))
	}
	if attr, ok := d.GetOk("metadata"); ok {
		if attr.(bool) {
			r.Metadata = upcloud.True
		} else {
			r.Metadata = upcloud.False
		}
	}
	if attr, ok := d.GetOk("cpu"); ok {
		r.CoreNumber = attr.(int)
	}
	if attr, ok := d.GetOk("mem"); ok {
		r.MemoryAmount = attr.(int)
	}
	if attr, ok := d.GetOk("timezone"); ok {
		r.TimeZone = attr.(string)
	}
	if attr, ok := d.GetOk("video_model"); ok {
		r.VideoModel = attr.(string)
	}
	if attr, ok := d.GetOk("nic_model"); ok {
		r.NICModel = attr.(string)
	}
	if attr, ok := d.GetOk("user_data"); ok {
		r.UserData = attr.(string)
	}
	if attr, ok := d.GetOk("plan"); ok {
		r.Plan = attr.(string)
	}
	if attr, ok := d.GetOk("boot_order"); ok {
		r.BootOrder = attr.(string)
	}
	if attr, ok := d.GetOk("simple_backup"); ok {
		simpleBackupAttrs := attr.(*schema.Set).List()[0].(map[string]interface{})
		r.SimpleBackup = buildSimpleBackupOpts(simpleBackupAttrs)
	}
	if login, ok := d.GetOk("login"); ok {
		loginOpts, deliveryMethod, err := buildLoginOpts(login)
		if err != nil {
			return nil, err
		}
		r.LoginUser = loginOpts
		r.PasswordDelivery = deliveryMethod
	}

	r.Host = d.Get("host").(int)

	if template, ok := d.GetOk("template.0"); ok {
		template := template.(map[string]interface{})
		if template["title"].(string) == "" {
			template["title"] = fmt.Sprintf("terraform-%s-disk", r.Hostname)
		}
		address := template["address"].(string)
		position := template["address_position"].(string)
		if position != "" {
			address += fmt.Sprintf(":%s", position)
		}
		serverStorageDevice := request.CreateServerStorageDevice{
			Action:    "clone",
			Address:   address,
			Encrypted: upcloud.FromBool(template["encrypt"].(bool)),
			Size:      template["size"].(int),
			Storage:   template["storage"].(string),
			Title:     template["title"].(string),
		}
		if attr, ok := d.GetOk("template.0.backup_rule.0"); ok {
			serverStorageDevice.BackupRule = storage.BackupRule(attr.(map[string]interface{}))
		}
		if source := template["storage"].(string); source != "" {
			// Assume template name is given and attempt map name to UUID
			if _, err := uuid.ParseUUID(source); err != nil {
				l, err := meta.(*service.Service).GetStorages(ctx, &request.GetStoragesRequest{
					Type: upcloud.StorageTypeTemplate,
				})
				if err != nil {
					return nil, err
				}
				for _, s := range l.Storages {
					if s.Title == source {
						source = s.UUID
						break
					}
				}
			}

			serverStorageDevice.Storage = source
		}
		r.StorageDevices = append(r.StorageDevices, serverStorageDevice)
	}

	if storageDevices, ok := d.GetOk("storage_devices"); ok {
		storageDevices := storageDevices.(*schema.Set)
		for _, storageDevice := range storageDevices.List() {
			storageDevice := storageDevice.(map[string]interface{})
			address := storageDevice["address"].(string)
			position := storageDevice["address_position"].(string)
			if position != "" {
				address += fmt.Sprintf(":%s", position)
			}
			r.StorageDevices = append(r.StorageDevices, request.CreateServerStorageDevice{
				Action:  "attach",
				Address: address,
				Type:    storageDevice["type"].(string),
				Storage: storageDevice["storage"].(string),
			})
		}
	}

	networking, err := buildNetworkOpts(d)
	if err != nil {
		return nil, err
	}

	r.Networking = &request.CreateServerNetworking{
		Interfaces: networking,
	}

	return r, nil
}

func buildLabels(m map[string]interface{}) *upcloud.LabelSlice {
	labels := upcloud.LabelSlice(utils.LabelsMapToSlice(m))
	return &labels
}

func buildSimpleBackupOpts(attrs map[string]interface{}) string {
	if backupTime, ok := attrs["time"]; ok {
		if plan, ok := attrs["plan"]; ok {
			return fmt.Sprintf("%s,%s", backupTime, plan)
		}
	}

	return "no"
}

func buildLoginOpts(v interface{}) (*request.LoginUser, string, error) {
	// Construct LoginUser struct from the schema
	r := &request.LoginUser{}
	e := v.(*schema.Set).List()[0]
	m := e.(map[string]interface{})

	// Set username as is
	r.Username = m["user"].(string)

	// Set 'create_password' to "yes" or "no" depending on the bool value.
	// Would be nice if the API would just get a standard bool str.
	createPassword := "no"
	b := m["create_password"].(bool)
	if b {
		createPassword = "yes"
	}
	r.CreatePassword = createPassword

	// Handle SSH keys one by one
	keys := make([]string, 0)
	for _, k := range m["keys"].([]interface{}) {
		key := k.(string)
		keys = append(keys, key)
	}
	r.SSHKeys = keys

	// Define password delivery method none/email/sms
	deliveryMethod := m["password_delivery"].(string)

	return r, deliveryMethod, nil
}

func buildNetworkOpts(d *schema.ResourceData) ([]request.CreateServerInterface, error) {
	ifaces := []request.CreateServerInterface{}

	niCount := d.Get("network_interface.#").(int)
	for i := 0; i < niCount; i++ {
		keyRoot := fmt.Sprintf("network_interface.%d.", i)
		iface := request.CreateServerInterface{
			Index: d.Get(keyRoot + "index").(int),
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family:  d.Get(keyRoot + "ip_address_family").(string),
					Address: d.Get(keyRoot + "ip_address").(string),
				},
			},
			Type: d.Get(keyRoot + "type").(string),
		}

		iface.SourceIPFiltering = upcloud.FromBool(d.Get(keyRoot + "source_ip_filtering").(bool))
		iface.Bootable = upcloud.FromBool(d.Get(keyRoot + "bootable").(bool))

		if v, ok := d.GetOk(keyRoot + "network"); ok {
			iface.Network = v.(string)
		}

		if additionalIPAddresses, ok := d.GetOk(keyRoot + "additional_ip_address"); ok {
			if iface.Type != upcloud.NetworkTypePrivate {
				return nil, fmt.Errorf("additional_ip_address can only be set for private network interfaces")
			}

			for _, v := range additionalIPAddresses.(*schema.Set).List() {
				ipAddress := v.(map[string]interface{})

				iface.IPAddresses = append(iface.IPAddresses, request.CreateServerIPAddress{
					Family:  ipAddress["ip_address_family"].(string),
					Address: ipAddress["ip_address"].(string),
				})
			}
		}

		ifaces = append(ifaces, iface)
	}

	return ifaces, nil
}
