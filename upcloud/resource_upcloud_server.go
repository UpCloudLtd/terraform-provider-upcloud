package upcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/server"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/storage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
)

const serverTitleLength int = 255

var errSubaccountCouldNotModifyTags = errors.New("creating and modifying tags is allowed only by main account. Subaccounts have access only to listing tags and tagged servers they are granted access to")

func resourceUpCloudServer() *schema.Resource {
	return &schema.Resource{
		Description:   "The UpCloud server resource allows the creation, update and deletion of a server.",
		CreateContext: resourceUpCloudServerCreate,
		ReadContext:   resourceUpCloudServerRead,
		UpdateContext: resourceUpCloudServerUpdate,
		DeleteContext: resourceUpCloudServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description:      "A valid domain name",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: serverValidateHostnameDiagFunc(1, 128),
			},
			"title": {
				Description:  "A short, informational description",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, serverTitleLength),
			},
			"zone": {
				Description: "The zone in which the server will be hosted",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
			"tags": {
				Description: "The server related tags",
				Type:        schema.TypeList,
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
				Type:        schema.TypeList,
				Description: "One or more blocks describing the network interfaces of the server.",
				Required:    true,
				ForceNew:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address_family": {
							Type:        schema.TypeString,
							Description: "The IP address type of this interface (one of `IPv4` or `IPv6`).",
							Optional:    true,
							ForceNew:    true,
							Default:     upcloud.IPAddressFamilyIPv4,
							ValidateDiagFunc: func(v interface{}, _ cty.Path) diag.Diagnostics {
								switch v.(string) {
								case upcloud.IPAddressFamilyIPv4, upcloud.IPAddressFamilyIPv6:
									return nil
								default:
									return diag.Diagnostics{diag.Diagnostic{
										Severity: diag.Error,
										Summary:  "'ip_address_family' has incorrect value",
										Detail: fmt.Sprintf(
											"'ip_address_family' must be one of %s or %s",
											upcloud.IPAddressFamilyIPv4,
											upcloud.IPAddressFamilyIPv6),
									}}
								}
							},
						},
						"ip_address": {
							Type:        schema.TypeString,
							Description: "The assigned IP address.",
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
						},
						"ip_address_floating": {
							Type:        schema.TypeBool,
							Description: "`true` is a floating IP address is attached.",
							Computed:    true,
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
							ForceNew:    true,
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
							ForceNew:    true,
							Optional:    true,
							Computed:    true,
						},
						"source_ip_filtering": {
							Type:        schema.TypeBool,
							Description: "`true` if source IP should be filtered.",
							ForceNew:    true,
							Optional:    true,
							Default:     true,
						},
						"bootable": {
							Type:        schema.TypeBool,
							Description: "`true` if this interface should be used for network booting.",
							ForceNew:    true,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"user_data": {
				Description: "Defines URL for a server setup script, or the script body itself",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"plan": {
				Description: "The pricing plan used for the server",
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
						"storage": {
							Description: "A valid storage UUID",
							Type:        schema.TypeString,
							Required:    true,
						},
						"address": {
							Description:  "The device address the storage will be attached to. Specify only the bus name (ide/scsi/virtio) to auto-select next available address from that bus.",
							Type:         schema.TypeString,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"scsi", "virtio", "ide"}, false),
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
					// compute a consistent hash for this TypeSet, mendatory
					m := v.(map[string]interface{})
					return schema.HashString(
						fmt.Sprintf("%s-%s", m["storage"].(string), m["address"].(string)),
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
							Description: "The device address the storage will be attached to. Specify only the bus name (ide/scsi/virtio) to auto-select next available address from that bus.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
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
							Description: "A valid storage UUID or template name",
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
						},
						"backup_rule": storage.BackupRuleSchema(),
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
							Description:  "The delivery method for the server’s root password",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "none",
							ValidateFunc: validation.StringInSlice([]string{"none", "email", "sms"}, false),
						},
					},
				},
			},
			"simple_backup": {
				Description: `Simple backup schedule configuration  
				The idea behind simple backups is to provide a simplified way of backing up *all* of the storages attached to a given server. 
				This means you cannot have simple backup set for a server, and then some individual backup_rules on the storages attached to said server. 
				Such configuration will throw an error during execution. This also apply to backup_rules set for server templates.  
				Also, due to how UpCloud API works with simple backups and how Terraform orders the update operations, 
				it is advised to never switch between simple_backup on the server and individual storages backup_rules in one apply.
				If you want to switch from using server simple backup to per-storage defined backup rules, 
				please first remove simple_backup block from a server, run 'terraform apply', 
				then add backup_rule to desired storages and run 'terraform apply' again.`,
				Type:          schema.TypeSet,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"template.0.backup_rule"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan": {
							Description:  "Simple backup plan. Accepted values: dailies, weeklies, monthlies.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"dailies", "weeklies", "monthlies"}, false),
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
		},
	}
}

func resourceUpCloudServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	tags, tagsExists := d.GetOk("tags")
	if tagsExists {
		if isSubaccount, err := isProviderAccountSubaccount(client); err != nil || isSubaccount {
			if err != nil {
				return diag.FromErr(err)
			}
			return diag.FromErr(errSubaccountCouldNotModifyTags)
		}
	}

	if err := serverValidatePlan(client, d.Get("plan").(string)); err != nil {
		return diag.FromErr(err)
	}

	if err := serverValidateZone(client, d.Get("zone").(string)); err != nil {
		return diag.FromErr(err)
	}

	r, err := server.BuildServerOpts(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("title"); ok {
		r.Title = d.Get("title").(string)
	} else {
		r.Title = serverDefaultTitleFromHostname(d.Get("hostname").(string))
	}

	serverDetails, err := client.CreateServer(r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serverDetails.UUID)
	log.Printf("[INFO] Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// set template id from the payload (if passed)
	if _, ok := d.GetOk("template.0"); ok {
		_ = d.Set("template", []map[string]interface{}{{
			"id":      serverDetails.StorageDevices[0].UUID,
			"storage": d.Get("template.0.storage"),
		}})
	}

	// add server tags
	if tagsExists {
		tags := utils.ExpandStrings(tags)
		if err := server.AddNewServerTags(client, serverDetails.UUID, tags); err != nil {
			return diag.FromErr(err)
		}

		if _, err := client.TagServer(&request.TagServerRequest{
			UUID: serverDetails.UUID,
			Tags: tags,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 25,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudServerRead(ctx, d, meta)
}

func resourceUpCloudServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == upcloudServerNotFoundErrorCode {
			diags = append(diags, diagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("hostname").(string)))
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	_ = d.Set("hostname", server.Hostname)
	if server.Title != serverDefaultTitleFromHostname(server.Hostname) {
		_ = d.Set("title", server.Title)
	} else {
		_ = d.Set("title", nil)
	}
	_ = d.Set("zone", server.Zone)
	_ = d.Set("cpu", server.CoreNumber)
	_ = d.Set("mem", server.MemoryAmount)
	_ = d.Set("metadata", server.Metadata.Bool())
	_ = d.Set("plan", server.Plan)

	// XXX: server.Tags returns an empty slice rather than nil when it's empty
	if len(server.Tags) > 0 {
		_ = d.Set("tags", server.Tags)
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

	networkInterfaces := []map[string]interface{}{}
	var connIP string
	for _, iface := range server.Networking.Interfaces {
		ni := make(map[string]interface{})
		ni["ip_address_family"] = iface.IPAddresses[0].Family
		ni["ip_address"] = iface.IPAddresses[0].Address
		if !iface.IPAddresses[0].Floating.Empty() {
			ni["ip_address_floating"] = iface.IPAddresses[0].Floating.Bool()
		}
		ni["mac_address"] = iface.MAC
		ni["network"] = iface.Network
		ni["type"] = iface.Type
		if !iface.Bootable.Empty() {
			ni["bootable"] = iface.Bootable.Bool()
		}
		if !iface.SourceIPFiltering.Empty() {
			ni["source_ip_filtering"] = iface.SourceIPFiltering.Bool()
		}

		networkInterfaces = append(networkInterfaces, ni)

		if iface.Type == upcloud.NetworkTypePublic &&
			iface.IPAddresses[0].Family == upcloud.IPAddressFamilyIPv4 {
			connIP = iface.IPAddresses[0].Address
		}
	}
	if len(networkInterfaces) > 0 {
		_ = d.Set("network_interface", networkInterfaces)
	}

	storageDevices := []interface{}{}
	log.Printf("[DEBUG] Configured storage devices in state: %+v", d.Get("storage_devices"))
	log.Printf("[DEBUG] Actual storage devices on server: %v", server.StorageDevices)
	for _, serverStorage := range server.StorageDevices {
		// the template is managed within the server
		if serverStorage.UUID == d.Get("template.0.id") {
			_ = d.Set("template", []map[string]interface{}{{
				"address": utils.StorageAddressFormat(serverStorage.Address),
				"id":      serverStorage.UUID,
				"size":    serverStorage.Size,
				"title":   serverStorage.Title,
				"storage": d.Get("template.0.storage"),
				"tier":    serverStorage.Tier,
				// NOTE: backupRule cannot be derived from server.storageDevices payload, will not sync if changed elsewhere
				"backup_rule": d.Get("template.0.backup_rule"),
			}})
		} else {
			storageDevices = append(storageDevices, map[string]interface{}{
				"address": utils.StorageAddressFormat(serverStorage.Address),
				"storage": serverStorage.UUID,
				"type":    serverStorage.Type,
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

func hasTemplateBackupRuleBeenReplacedWithSimpleBackups(d *schema.ResourceData) bool {
	if !d.HasChange("simple_backup") || !d.HasChange("template.0.backup_rule") {
		return false
	}

	sb, sbOk := d.GetOk("simple_backup")
	if !sbOk {
		return false
	}

	simpleBackup := sb.(*schema.Set).List()[0].(map[string]interface{})
	if simpleBackup["interval"] == "" {
		return false
	}

	tbr, tbrOk := d.GetOk("template.0.backup_rule.0")
	templateBackupRule := tbr.(map[string]interface{})
	if tbrOk && templateBackupRule["interval"] != "" {
		return false
	}

	return true
}

func resourceUpCloudServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	planHasChange := d.HasChange("plan")
	if planHasChange {
		if err := serverValidatePlan(client, d.Get("plan").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	tagsHasChange := d.HasChange("tags")
	if tagsHasChange {
		if isSubaccount, err := isProviderAccountSubaccount(client); err != nil || isSubaccount {
			if err != nil {
				return diag.FromErr(err)
			}
			return diag.FromErr(errSubaccountCouldNotModifyTags)
		}
	}

	serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// Stop the server if the requested changes require it
	if d.HasChanges("cpu", "mem", "template.0.size", "storage_devices") || planHasChange {
		err := server.VerifyServerStopped(request.StopServerRequest{
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

	if attr, ok := d.GetOk("title"); ok {
		r.Title = attr.(string)
	} else {
		r.Title = serverDefaultTitleFromHostname(d.Get("hostname").(string))
	}

	r.Metadata = upcloud.FromBool(d.Get("metadata").(bool))

	if d.Get("firewall").(bool) {
		r.Firewall = "on"
	} else {
		r.Firewall = "off"
	}

	if d.HasChange(("simple_backup")) {
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

				tmpl, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{UUID: templateID})
				if err != nil {
					return diag.FromErr(err)
				}

				if tmpl.BackupRule != nil && tmpl.BackupRule.Interval != "" {
					r := &request.ModifyStorageRequest{
						UUID:       templateID,
						BackupRule: &upcloud.BackupRule{},
					}

					if _, err := client.ModifyStorage(r); err != nil {
						return diag.FromErr(err)
					}
				}
			}

			simpleBackupAttrs := sb.(*schema.Set).List()[0].(map[string]interface{})
			r.SimpleBackup = server.BuildSimpleBackupOpts(simpleBackupAttrs)
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

	if _, err := client.ModifyServer(r); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("tags"); ok && tagsHasChange {
		oldTags, newTags := d.GetChange("tags")
		if err := server.UpdateServerTags(
			client, d.Id(),
			utils.ExpandStrings(oldTags), utils.ExpandStrings(newTags)); err != nil {
			return diag.FromErr(err)
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

		if _, err := client.ModifyStorage(r); err != nil {
			return diag.FromErr(err)
		}
	}

	// should reattach if address changed
	if d.HasChange("template.0.address") {
		o, n := d.GetChange("template.0.address")
		if _, err := client.DetachStorage(&request.DetachStorageRequest{
			ServerUUID: d.Id(),
			Address:    utils.StorageAddressFormat(o.(string)),
		}); err != nil {
			return diag.FromErr(err)
		}
		if _, err := client.AttachStorage(&request.AttachStorageRequest{
			Address:     utils.StorageAddressFormat(n.(string)),
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
			if _, err := client.DetachStorage(&request.DetachStorageRequest{
				ServerUUID: d.Id(),
				Address:    serverStorageDevice.Address,
			}); err != nil {
				return diag.FromErr(err)
			}

			// Remove backup rule from the detached storage, if it was a result of simple backup setting
			if _, ok := d.GetOk("simple_backup"); ok {
				if _, err := client.ModifyStorage(&request.ModifyStorageRequest{
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
			if _, err := client.AttachStorage(&request.AttachStorageRequest{
				ServerUUID:  d.Id(),
				Address:     utils.StorageAddressFormat(storageDevice["address"].(string)),
				StorageUUID: storageDevice["storage"].(string),
				Type:        storageDevice["type"].(string),
			}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err := server.VerifyServerStarted(request.StartServerRequest{UUID: d.Id(), Host: d.Get("host").(int)}, meta); err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudServerRead(ctx, d, meta)
}

func resourceUpCloudServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	// Verify server is stopped before deletion
	if err := server.VerifyServerStopped(request.StopServerRequest{UUID: d.Id()}, meta); err != nil {
		return diag.FromErr(err)
	}
	// Delete server
	deleteServerRequest := &request.DeleteServerRequest{
		UUID: d.Id(),
	}
	log.Printf("[INFO] Deleting server (server UUID: %s)", d.Id())
	err := client.DeleteServer(deleteServerRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete server root disk
	if template, ok := d.GetOk("template.0"); ok {
		template := template.(map[string]interface{})
		deleteStorageRequest := &request.DeleteStorageRequest{
			UUID: template["id"].(string),
		}
		log.Printf("[INFO] Deleting server storage (storage UUID: %s)", deleteStorageRequest.UUID)
		err = client.DeleteStorage(deleteStorageRequest)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func serverDefaultTitleFromHostname(hostname string) string {
	const suffix string = " (managed by terraform)"
	if len(hostname)+len(suffix) > serverTitleLength {
		hostname = fmt.Sprintf("%s…", hostname[:serverTitleLength-len(suffix)-1])
	}
	return fmt.Sprintf("%s%s", hostname, suffix)
}

func serverValidateHostnameDiagFunc(min, max int) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		val, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Bad type",
				Detail:        "expected type to be string",
				AttributePath: path,
			})
			return diags
		}

		if len(val) < min || len(val) > max {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Hostname length validation failed",
				Detail:        fmt.Sprintf("expected hostname length to be in the range (%d - %d), got %d", min, max, len(val)),
				AttributePath: path,
			})
			return diags
		}

		if err := serverValidateHostname(val); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Hostname validation failed",
				Detail:        err.Error(),
				AttributePath: path,
			})
		}

		return diags
	}
}

// Validate server hostname
//
// hostname(7): Each element of the hostname must be from 1 to 63 characters long
// and the entire hostname, including the dots, can be at most 253 characters long.
// Valid characters for hostnames are ASCII(7) letters from a to z, the digits from 0 to 9, and the hyphen (-).
// A hostname may not start with a hyphen.
//
// Modified version of isDomainName function from Go's net package (https://pkg.go.dev/net)
func serverValidateHostname(hostname string) error {
	const (
		minLen      int = 1
		maxLen      int = 253
		labelMaxLen int = 63
	)
	l := len(hostname)

	if l > maxLen || l < minLen {
		return fmt.Errorf("%s length %d is not in the range %d - %d", hostname, l, minLen, maxLen)
	}

	if hostname[0] == '.' || hostname[0] == '-' {
		return fmt.Errorf("%s starts with dot or hyphen", hostname)
	}

	if hostname[l-1] == '.' || hostname[l-1] == '-' {
		return fmt.Errorf("%s ends with dot or hyphen", hostname)
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen (either one is required)
	labelLen := 0

	for i := 0; i < l; i++ {
		c := hostname[i]
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			labelLen++
		case '0' <= c && c <= '9':
			labelLen++
		case c == '-':
			if last == '.' {
				return fmt.Errorf("'%s' character before hyphen cannot be dot", hostname[0:i+1])
			}
			labelLen++
			nonNumeric = true
		case c == '.':
			if last == '.' || last == '-' {
				return fmt.Errorf("'%s' character before dot cannot be dot or hyphen", hostname[0:i+1])
			}
			if labelLen > labelMaxLen || labelLen == 0 {
				return fmt.Errorf("'%s' label is not in the range %d - %d", hostname[0:i+1], minLen, labelMaxLen)
			}
			labelLen = 0
		default:
			return fmt.Errorf("%s contains illegal characters", hostname)
		}
		last = c
	}

	if labelLen > labelMaxLen {
		return fmt.Errorf("%s label is not in the range %d - %d", hostname, minLen, labelMaxLen)
	}

	if !nonNumeric {
		return fmt.Errorf("%s contains only numeric labels", hostname)
	}

	return nil
}

func serverValidatePlan(service *service.Service, plan string) error {
	if plan == "" {
		return nil
	}
	plans, err := service.GetPlans()
	if err != nil {
		return err
	}
	availablePlans := make([]string, 0)
	for _, p := range plans.Plans {
		if p.Name == plan {
			return nil
		}
		availablePlans = append(availablePlans, p.Name)
	}
	return fmt.Errorf("expected plan to be one of [%s], got %s", strings.Join(availablePlans, ", "), plan)
}

func serverValidateZone(service *service.Service, zone string) error {
	zones, err := service.GetZones()
	if err != nil {
		return err
	}
	availableZones := make([]string, 0)
	for _, z := range zones.Zones {
		if z.ID == zone {
			return nil
		}
		availableZones = append(availableZones, z.ID)
	}
	return fmt.Errorf("expected zone to be one of [%s], got %s", strings.Join(availableZones, ", "), zone)
}
