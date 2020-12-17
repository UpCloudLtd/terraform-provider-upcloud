package upcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"cpu": {
				Description:   "The number of CPU for the server",
				Type:          schema.TypeInt,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"plan"},
			},
			"mem": {
				Description:   "The size of memory for the server",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"plan"},
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
							Computed:    true,
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
							Description: "The device address the storage will be attached to. Specify only the bus name (ide/scsi/virtio) to auto-select next available address from that bus.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
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
			},
			"template": {
				Description: "",
				Type:        schema.TypeList,
				// NOTE: might want to make this optional
				Required: true,
				MaxItems: 1,
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
							ForceNew:    true,
							Optional:    true,
						},
						"size": {
							Description: "The size of the storage in gigabytes",
							Type:        schema.TypeInt,
							// TODO: update go-api to omit zero value from the payload and make this optional
							Required:     true,
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
						"backup_rule": backupRuleSchema(),
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
							Required:    true,
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

	r, err := buildServerOpts(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	server, err := client.CreateServer(r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(server.UUID)
	log.Printf("[INFO] Server %s with UUID %s created", server.Title, server.UUID)

	server, err = client.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         server.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 25,
	})

	// set template id from the payload (if passed)
	if _, ok := d.GetOk("template.0"); ok {
		d.Set("template", []map[string]interface{}{{
			"id":      server.StorageDevices[0].UUID,
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
	d.Set("hostname", server.Hostname)
	d.Set("title", server.Title)
	d.Set("zone", server.Zone)
	d.Set("cpu", server.CoreNumber)
	d.Set("mem", server.MemoryAmount)

	networkInterfaces := []map[string]interface{}{}
	var connIP string
	for _, iface := range server.Networking.Interfaces {
		ni := make(map[string]interface{}, 0)
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
		d.Set("network_interface", networkInterfaces)
	}

	storageDevices := []interface{}{}
	log.Printf("[DEBUG] Configured storage devices in state: %+v", d.Get("storage_devices"))
	log.Printf("[DEBUG] Actual storage devices on server: %v", server.StorageDevices)
	for _, serverStorage := range server.StorageDevices {
		// the template is managed within the server
		if serverStorage.UUID == d.Get("template.0.id") {
			d.Set("template", []map[string]interface{}{{
				"address": serverStorage.Address,
				"id":      serverStorage.UUID,
				"size":    serverStorage.Size,
				"title":   serverStorage.Title,
				"storage": d.Get("template.0.storage"),
				// FIXME: backupRule cannot be derived from server.storageDevices payload, will not sync if changed elsewhere
				"backup_rule": d.Get("template.0.backup_rule"),
				// TODO: add when go-api updated ... "tier":   serverStorage.Tier,
			}})
		} else {
			storageDevices = append(storageDevices, map[string]interface{}{
				"address": serverStorage.Address,
				"storage": serverStorage.UUID,
				"type":    serverStorage.Type,
			})
		}
	}
	d.Set("storage_devices", storageDevices)

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

	if err := verifyServerStopped(d.Id(), meta); err != nil {
		return diag.FromErr(err)
	}

	r := &request.ModifyServerRequest{
		UUID: d.Id(),
	}

	if d.Get("firewall").(bool) {
		r.Firewall = "on"
	} else {
		r.Firewall = "off"
	}

	if plan, ok := d.GetOk("plan"); ok {
		r.Plan = plan.(string)
	} else {
		r.CoreNumber = d.Get("cpu").(int)
		r.MemoryAmount = d.Get("mem").(int)
	}
	r.Hostname = d.Get("hostname").(string)

	if _, err := client.ModifyServer(r); err != nil {
		return diag.FromErr(err)
	}

	// handle the template
	if d.HasChanges("template.0.title", "template.0.size", "template.0.backup_rule") {
		template := d.Get("template.0").(map[string]interface{})
		if _, err := client.ModifyStorage(&request.ModifyStorageRequest{
			UUID:  template["id"].(string),
			Size:  template["size"].(int),
			Title: template["title"].(string),
			// TODO: handle backup_rule
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	// handle the other storage devices
	if d.HasChange("storage_devices") {
		o, n := d.GetChange("storage_devices")

		// detach the devices that should be detached or sould be re-attached with different parameters
		for _, storage_device := range o.(*schema.Set).Difference(n.(*schema.Set)).List() {
			if _, err := client.DetachStorage(&request.DetachStorageRequest{
				ServerUUID: d.Id(),
				Address:    storage_device.(map[string]interface{})["address"].(string),
			}); err != nil {
				return diag.FromErr(err)
			}
		}
		// attach the storages that are new or have changed
		for _, storage_device := range n.(*schema.Set).Difference(o.(*schema.Set)).List() {
			storage_device := storage_device.(map[string]interface{})
			if _, err := client.AttachStorage(&request.AttachStorageRequest{
				ServerUUID:  d.Id(),
				Address:     storage_device["address"].(string),
				StorageUUID: storage_device["storage"].(string),
				Type:        storage_device["type"].(string),
			}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err := verifyServerStarted(d.Id(), meta); err != nil {
		return diag.FromErr(err)
	}
	return resourceUpCloudServerRead(ctx, d, meta)
}

func resourceUpCloudServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	var diags diag.Diagnostics

	// Verify server is stopped before deletion
	if err := verifyServerStopped(d.Id(), meta); err != nil {
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

func buildServerOpts(d *schema.ResourceData, meta interface{}) (*request.CreateServerRequest, error) {
	r := &request.CreateServerRequest{
		Zone:     d.Get("zone").(string),
		Hostname: d.Get("hostname").(string),
		Title:    fmt.Sprintf("%s (managed by terraform)", d.Get("hostname").(string)),
	}

	if attr, ok := d.GetOk("firewall"); ok {
		if attr.(bool) {
			r.Firewall = "on"
		} else {
			r.Firewall = "off"
		}
	}
	if attr, ok := d.GetOk("cpu"); ok {
		r.CoreNumber = attr.(int)
	}
	if attr, ok := d.GetOk("mem"); ok {
		r.MemoryAmount = attr.(int)
	}
	if attr, ok := d.GetOk("user_data"); ok {
		r.UserData = attr.(string)
	}
	if attr, ok := d.GetOk("plan"); ok {
		r.Plan = attr.(string)
	}
	if login, ok := d.GetOk("login"); ok {
		loginOpts, deliveryMethod, err := buildLoginOpts(login, meta)
		if err != nil {
			return nil, err
		}
		r.LoginUser = loginOpts
		r.PasswordDelivery = deliveryMethod
	}

	if template, ok := d.GetOk("template.0"); ok {
		template := template.(map[string]interface{})
		if template["title"].(string) == "" {
			template["title"] = fmt.Sprintf("terraform-%s-disk", r.Hostname)
		}
		r.StorageDevices = append(
			r.StorageDevices,
			request.CreateServerStorageDevice{
				Action:  "clone",
				Address: template["address"].(string),
				Size:    template["size"].(int),
				Storage: template["storage"].(string),
				Title:   template["title"].(string),
			},
		)
		// TODO: handle backup_rule
	}

	if storage_devices, ok := d.GetOk("storage_devices"); ok {
		storage_devices := storage_devices.(*schema.Set)
		for _, storage_device := range storage_devices.List() {
			storage_device := storage_device.(map[string]interface{})
			r.StorageDevices = append(r.StorageDevices, request.CreateServerStorageDevice{
				Action:  "attach",
				Address: storage_device["address"].(string),
				Type:    storage_device["type"].(string),
				Storage: storage_device["storage"].(string),
			})
		}
	}

	networking, err := buildNetworkOpts(d, meta)
	if err != nil {
		return nil, err
	}

	r.Networking = &request.CreateServerNetworking{
		Interfaces: networking,
	}

	return r, nil
}

func buildNetworkOpts(d *schema.ResourceData, meta interface{}) ([]request.CreateServerInterface, error) {
	ifaces := []request.CreateServerInterface{}

	niCount := d.Get("network_interface.#").(int)
	for i := 0; i < niCount; i++ {
		keyRoot := fmt.Sprintf("network_interface.%d.", i)

		iface := request.CreateServerInterface{
			IPAddresses: []request.CreateServerIPAddress{
				{
					Family: d.Get(keyRoot + "ip_address_family").(string),
				},
			},
			Type: d.Get(keyRoot + "type").(string),
		}

		iface.SourceIPFiltering = upcloud.FromBool(d.Get(keyRoot + "source_ip_filtering").(bool))
		iface.Bootable = upcloud.FromBool(d.Get(keyRoot + "bootable").(bool))

		if v, ok := d.GetOk(keyRoot + "network"); ok {
			iface.Network = v.(string)
		}

		ifaces = append(ifaces, iface)
	}

	return ifaces, nil
}

func buildLoginOpts(v interface{}, meta interface{}) (*request.LoginUser, string, error) {
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

func verifyServerStopped(id string, meta interface{}) error {
	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: id,
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStopped {
		// Soft stop with 2 minute timeout, after which hard stop occurs
		stopRequest := &request.StopServerRequest{
			UUID:     id,
			StopType: "soft",
			Timeout:  time.Minute * 2,
		}
		log.Printf("[INFO] Stopping server (server UUID: %s)", id)
		_, err := client.StopServer(stopRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         id,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyServerStarted(id string, meta interface{}) error {
	client := meta.(*service.Service)
	// Get current server state
	r := &request.GetServerDetailsRequest{
		UUID: id,
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}
	if server.State != upcloud.ServerStateStarted {
		startRequest := &request.StartServerRequest{
			UUID:    id,
			Timeout: time.Minute * 2,
		}
		log.Printf("[INFO] Starting server (server UUID: %s)", id)
		_, err := client.StartServer(startRequest)
		if err != nil {
			return err
		}
		_, err = client.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         id,
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      time.Minute * 5,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
