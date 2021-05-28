package upcloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/server"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/storage"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudServerCreate,
		ReadContext:   resourceUpCloudServerRead,
		UpdateContext: resourceUpCloudServerUpdate,
		DeleteContext: resourceUpCloudServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"hostname": {
				Description:  "A valid domain name",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"title": {
				Description: "A short, informational description",
				Type:        schema.TypeString,
				Computed:    true,
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
				Description: "Block describing the preconfigured operating system",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
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
		},
	}
}

func resourceUpCloudServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	r, err := server.BuildServerOpts(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	serverDetails, err := client.CreateServer(r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serverDetails.UUID)
	log.Printf("[INFO] Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

	// add server tags
	if _, ok := d.GetOk("tags"); ok {
		tags := utils.ExpandStrings(d.Get("tags"))
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

	serverDetails, err = client.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 25,
	})

	// set template id from the payload (if passed)
	if _, ok := d.GetOk("template.0"); ok {
		_ = d.Set("template", []map[string]interface{}{{
			"id":      serverDetails.StorageDevices[0].UUID,
			"storage": d.Get("template.0.storage"),
		}})
	}

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
		return diag.FromErr(err)
	}
	_ = d.Set("hostname", server.Hostname)
	_ = d.Set("title", server.Title)
	_ = d.Set("zone", server.Zone)
	_ = d.Set("cpu", server.CoreNumber)
	_ = d.Set("mem", server.MemoryAmount)
	_ = d.Set("metadata", server.Metadata.Bool())
	_ = d.Set("plan", server.Plan)
	_ = d.Set("tags", server.Tags)
	if server.Firewall == "on" {
		_ = d.Set("firewall", true)
	} else {
		_ = d.Set("firewall", false)
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

func resourceUpCloudServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{
		UUID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	r := &request.ModifyServerRequest{
		UUID: d.Id(),
	}

	r.Hostname = d.Get("hostname").(string)
	r.Title = fmt.Sprintf("%s (managed by terraform)", r.Hostname)
	r.Metadata = upcloud.FromBool(d.Get("metadata").(bool))

	if d.Get("firewall").(bool) {
		r.Firewall = "on"
	} else {
		r.Firewall = "off"
	}
	if _, ok := d.GetOk("tags"); ok {
		if d.HasChange("tags") {
			oldTags, newTags := d.GetChange("tags")

			if err := server.UpdateServerTags(
				client, d.Id(),
				utils.ExpandStrings(oldTags), utils.ExpandStrings(newTags)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// handle changes that need reboot
	if d.HasChanges("plan", "cpu", "mem", "template", "storage_devices") {
		if err := server.VerifyServerStopped(
			request.StopServerRequest{
				UUID: d.Id(),
			},
			meta,
		); err != nil {
			return diag.FromErr(err)
		}

		if plan, ok := d.GetOk("plan"); ok {
			r.Plan = plan.(string)
		} else {
			r.CoreNumber = d.Get("cpu").(int)
			r.MemoryAmount = d.Get("mem").(int)
			r.Plan = "custom"
		}

		// handle the template
		if d.HasChanges("template.0.title", "template.0.size", "template.0.backup_rule") {
			template := d.Get("template.0").(map[string]interface{})
			if _, err := client.ModifyStorage(&request.ModifyStorageRequest{
				UUID:  template["id"].(string),
				Size:  template["size"].(int),
				Title: template["title"].(string),
				BackupRule: storage.BackupRule(
					d.Get("template.0.backup_rule.0").(map[string]interface{}),
				),
			}); err != nil {
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
	}

	if _, err := client.ModifyServer(r); err != nil {
		return diag.FromErr(err)
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
