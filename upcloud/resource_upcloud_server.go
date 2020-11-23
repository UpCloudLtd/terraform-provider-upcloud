package upcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	uuid "github.com/hashicorp/go-uuid"
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
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MinItems:    1,
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
						"action": {
							Description:  "The method used to create or attach the specified storage",
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"create", "clone", "attach"}, false),
						},
						"size": {
							Description:  "The size of the storage in gigabytes",
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      -1,
							ValidateFunc: validation.IntBetween(10, 2048),
						},
						"tier": {
							Description:  "The storage tier to use",
							Type:         schema.TypeString,
							Default:      "hdd",
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"hdd", "maxiops"}, false),
						},
						"title": {
							Description:  "A short, informative description",
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 64),
						},
						"storage": {
							Description: "A valid storage UUID or template name",
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
						},
						"type": {
							Description:  "The device type the storage will be attached as",
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							Default:      "disk",
							ValidateFunc: validation.StringInSlice([]string{"disk", "cdrom"}, false),
						},
						"backup_rule": {
							Description: "The criteria to backup the storage",
							Type:        schema.TypeSet,
							MaxItems:    1,
							ForceNew:    true,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"interval": {
										Description: "The weekday when the backup is created",
										Type:        schema.TypeString,
										ForceNew:    true,
										Required:    true,
									},
									"time": {
										Description: "The time of day when the backup is created",
										Type:        schema.TypeString,
										ForceNew:    true,
										Required:    true,
									},
									"retention": {
										Description: "The number of days before a backup is automatically deleted",
										Type:        schema.TypeString,
										ForceNew:    true,
										Required:    true,
									},
								},
							},
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
	if err != nil {
		return diag.FromErr(err)
	}

	err = buildAfterServerCreationOps(d, client)
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

	storageDevices := d.Get("storage_devices").([]interface{})
	log.Printf("[DEBUG] Configured storage devices in state: %v", storageDevices)
	log.Printf("[DEBUG] Actual storage devices on server: %v", server.StorageDevices)
	for i, storageDevice := range storageDevices {
		storageDevice := storageDevice.(map[string]interface{})
		storageDevice["id"] = server.StorageDevices[i].UUID
		storageDevice["address"] = server.StorageDevices[i].Address
		storageDevice["title"] = server.StorageDevices[i].Title
		storageDevice["size"] = server.StorageDevices[i].Size
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

	storageDevices := d.Get("storage_devices").([]interface{})
	for _, storageDevice := range storageDevices {
		// Delete server root disk
		storageDevice := storageDevice.(map[string]interface{})
		id := storageDevice["id"].(string)
		action := storageDevice["action"].(string)
		if action != request.CreateServerStorageDeviceActionAttach {
			deleteStorageRequest := &request.DeleteStorageRequest{
				UUID: id,
			}
			log.Printf("[INFO] Deleting server storage (storage UUID: %s)", id)
			err = client.DeleteStorage(deleteStorageRequest)
			if err != nil {
				return diag.FromErr(err)
			}
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

	storageDevices := d.Get("storage_devices").([]interface{})
	storageOpts, err := buildStorageOpts(storageDevices, meta, d.Get("hostname").(string), d.Get("zone").(string))
	if err != nil {
		return nil, err
	}
	r.StorageDevices = storageOpts

	networking, err := buildNetworkOpts(d, meta)
	if err != nil {
		return nil, err
	}

	r.Networking = &request.CreateServerNetworking{
		Interfaces: networking,
	}

	return r, nil
}

func buildAfterServerCreationOps(d *schema.ResourceData, meta interface{}) error {
	/*
		Some of the operations such as backup_rule for storage device can only be done after
		the server creation.
	*/

	err := buildStorageBackupRuleOps(d, meta)
	if err != nil {
		return err
	}

	return nil
}

func buildStorageBackupRuleOps(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*service.Service)

	storageDevices := d.Get("storage_devices").([]interface{})

	r := &request.GetServerDetailsRequest{
		UUID: d.Id(),
	}
	server, err := client.GetServerDetails(r)
	if err != nil {
		return err
	}

	for i, storageDevice := range storageDevices {
		storageDevice := storageDevice.(map[string]interface{})

		if backupRule := storageDevice["backup_rule"].(*schema.Set).List(); backupRule != nil && len(backupRule) != 0 {

			for _, br := range backupRule {
				mBr := br.(map[string]interface{})

				retentionValue, err := strconv.Atoi(mBr["retention"].(string))

				if err != nil {
					diag.FromErr(err)
				}

				modifyStorage := &request.ModifyStorageRequest{
					UUID: server.StorageDevices[i].UUID,
				}

				modifyStorage.BackupRule = &upcloud.BackupRule{
					Interval:  mBr["interval"].(string),
					Time:      mBr["time"].(string),
					Retention: retentionValue,
				}

				client.ModifyStorage(modifyStorage)
			}
		}
	}

	return nil
}

func buildStorage(storageDevice map[string]interface{}, i int, meta interface{}, hostname, zone string) (*request.CreateServerStorageDevice, error) {
	osDisk := request.CreateServerStorageDevice{}

	if source := storageDevice["storage"].(string); source != "" {
		_, err := uuid.ParseUUID(source)
		// Assume template name is given and map name to UUID
		if err != nil {
			client := meta.(*service.Service)
			r := &request.GetStoragesRequest{
				Type: upcloud.StorageTypeTemplate,
			}
			l, err := client.GetStorages(r)
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

		osDisk.Storage = source
	}

	// Set size or use the one defined by target template
	if size := storageDevice["size"]; size != -1 {
		osDisk.Size = size.(int)
	}

	// Autogenerate disk title
	if title := storageDevice["title"].(string); title != "" {
		osDisk.Title = title
	} else {
		osDisk.Title = fmt.Sprintf("terraform-%s-disk-%d", hostname, i)
	}

	// Set disk tier or use the one defined by target template
	if tier := storageDevice["tier"]; tier != "" {
		osDisk.Tier = tier.(string)
	}

	if storageType := storageDevice["type"].(string); storageType != "" {
		osDisk.Type = storageType
	}

	if address := storageDevice["address"].(string); address != "" {
		osDisk.Address = address
	}

	osDisk.Action = storageDevice["action"].(string)

	log.Printf("[DEBUG] Disk: %v", osDisk)

	return &osDisk, nil
}

func buildStorageOpts(storageDevices []interface{}, meta interface{}, hostname, zone string) ([]request.CreateServerStorageDevice, error) {
	storageCfg := make([]request.CreateServerStorageDevice, 0)
	for i, storageDevice := range storageDevices {
		storageDevice, err := buildStorage(storageDevice.(map[string]interface{}), i, meta, hostname, zone)

		if err != nil {
			return nil, err
		}

		storageCfg = append(storageCfg, *storageDevice)
	}

	return storageCfg, nil
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
