package managedobjectstorage

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceManagedObjectStorage() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents an UpCloud Managed Object Storage instance, which provides S3 compatible storage.",
		CreateContext: resourceManagedObjectStorageCreate,
		ReadContext:   resourceManagedObjectStorageRead,
		UpdateContext: resourceManagedObjectStorageUpdate,
		DeleteContext: resourceManagedObjectStorageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"configured_status": {
				Description: "Service status managed by the end user.",
				Required:    true,
				Type:        schema.TypeString,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(upcloud.ManagedObjectStorageConfiguredStatusStarted),
						string(upcloud.ManagedObjectStorageConfiguredStatusStopped),
					},
					false,
				),
			},
			"created_at": {
				Description: "Creation time.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"endpoint": {
				Description: "Endpoints for accessing the Managed Object Storage service.",
				Computed:    true,
				Type:        schema.TypeSet,
				Elem:        schemaEndpoint(),
			},
			"labels": utils.LabelsSchema("managed object storage"),
			"network": {
				Description: "Attached networks from where object storage can be used. Private networks must reside in object storage region.",
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        schemaNetwork(),
			},
			"operational_state": {
				Description: "Operational state of the Managed Object Storage service.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"region": {
				Description: "Region in which the service will be hosted, see `upcloud_managed_object_storage_regions` data source.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"updated_at": {
				Description: "Creation time.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"users": {
				Description: "List of UpCloud API users allowed to use object storage. Valid values include current account and it's sub-accounts. See `upcloud_managed_object_storage_user_access_key` for managing access keys.",
				Optional:    true,
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func schemaEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: "Endpoint",
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Description: "Domain name of the endpoint.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"type": {
				Description: "Type of the endpoint (`private` / `public`).",
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func schemaNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "Network",
		Schema: map[string]*schema.Schema{
			"family": {
				Description: "Network family. IPv6 currently not supported.",
				Required:    true,
				Type:        schema.TypeString,
				ValidateFunc: validation.StringInSlice(
					[]string{
						upcloud.IPAddressFamilyIPv4,
						// IPv6 currently not supported
					},
					false,
				),
			},
			"name": {
				Description: "Network name. Must be unique within the service.",
				Required:    true,
				Type:        schema.TypeString,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), ""),
				),
			},
			"type": {
				Description: "Network type.",
				Required:    true,
				Type:        schema.TypeString,
				ValidateFunc: validation.StringInSlice(
					[]string{
						"private",
						"public",
					},
					false,
				),
			},
			"uuid": {
				Description: "Private network uuid. For public networks the field should be omitted.",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceManagedObjectStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.CreateManagedObjectStorageRequest{
		ConfiguredStatus: upcloud.ManagedObjectStorageConfiguredStatus(d.Get("configured_status").(string)),
		Region:           d.Get("region").(string),
	}

	if v, ok := d.GetOk("labels"); ok {
		req.Labels = utils.LabelsMapToSlice(v.(map[string]interface{}))
	}

	networks, err := networksFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Networks = networks

	userReqs := make([]request.ManagedObjectStorageUser, 0)
	users, ok := d.GetOk("users")
	if ok {
		for _, user := range users.(*schema.Set).List() {
			userReqs = append(req.Users, request.ManagedObjectStorageUser{
				Username: user.(string),
			})
		}
	}
	req.Users = userReqs

	storage, err := svc.CreateManagedObjectStorage(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(storage.UUID)

	waitReq := &request.WaitForManagedObjectStorageOperationalStateRequest{
		DesiredState: upcloud.ManagedObjectStorageOperationalStateRunning,
		Timeout:      time.Minute * 20,
		UUID:         storage.UUID,
	}

	if storage.ConfiguredStatus == upcloud.ManagedObjectStorageConfiguredStatusStopped {
		waitReq.DesiredState = upcloud.ManagedObjectStorageOperationalStateStopped
	}

	storage, err = svc.WaitForManagedObjectStorageOperationalState(ctx, waitReq)
	if err != nil {
		return diag.FromErr(err)
	}

	return append(diags, setManagedObjectStorageData(d, storage)...)
}

func resourceManagedObjectStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	svc := meta.(*service.Service)

	storage, err := svc.GetManagedObjectStorage(ctx, &request.GetManagedObjectStorageRequest{UUID: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageData(d, storage)
}

func resourceManagedObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.ModifyManagedObjectStorageRequest{
		UUID: d.Id(),
	}

	if d.HasChange("configured_status") {
		configuredStatus := upcloud.ManagedObjectStorageConfiguredStatus(d.Get("configured_status").(string))
		req.ConfiguredStatus = &configuredStatus
	}

	if d.HasChange("labels") {
		labels := utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{}))
		req.Labels = &labels
	}

	if d.HasChange("network") {
		networks, err := networksFromResourceData(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req.Networks = &networks
	}

	if d.HasChange("users") {
		userReqs := make([]request.ManagedObjectStorageUser, 0)
		users := d.Get("users")
		for _, user := range users.(*schema.Set).List() {
			r := request.ManagedObjectStorageUser{
				Username: user.(string),
			}
			userReqs = append(userReqs, r)
		}
		req.Users = &userReqs
	}
	storage, err := svc.ModifyManagedObjectStorage(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	waitReq := &request.WaitForManagedObjectStorageOperationalStateRequest{
		DesiredState: upcloud.ManagedObjectStorageOperationalStateRunning,
		Timeout:      time.Minute * 20,
		UUID:         storage.UUID,
	}

	if storage.ConfiguredStatus == upcloud.ManagedObjectStorageConfiguredStatusStopped {
		waitReq.DesiredState = upcloud.ManagedObjectStorageOperationalStateStopped
	}

	storage, err = svc.WaitForManagedObjectStorageOperationalState(ctx, waitReq)
	if err != nil {
		return diag.FromErr(err)
	}

	return setManagedObjectStorageData(d, storage)
}

func resourceManagedObjectStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)
	err := svc.DeleteManagedObjectStorage(ctx, &request.DeleteManagedObjectStorageRequest{UUID: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	err = svc.WaitForManagedObjectStorageDeletion(ctx, &request.WaitForManagedObjectStorageDeletionRequest{
		Timeout: time.Minute * 20,
		UUID:    d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func networksFromResourceData(d *schema.ResourceData) ([]upcloud.ManagedObjectStorageNetwork, error) {
	req := make([]upcloud.ManagedObjectStorageNetwork, 0)
	if networks, ok := d.GetOk("network"); ok {
		for i, network := range networks.(*schema.Set).List() {
			n := network.(map[string]interface{})
			r := upcloud.ManagedObjectStorageNetwork{
				Name:   n["name"].(string),
				Type:   n["type"].(string),
				Family: n["family"].(string),
				UUID:   upcloud.StringPtr(n["uuid"].(string)),
			}
			if r.Type == "private" && *r.UUID == "" {
				return req, fmt.Errorf("private network (#%d) UUID is required", i)
			}
			if r.Type == "public" && *r.UUID != "" {
				return req, fmt.Errorf("setting UUID for a public network (#%d) is not supported", i)
			}
			req = append(req, r)
		}
	}

	return req, nil
}

func setManagedObjectStorageData(d *schema.ResourceData, storage *upcloud.ManagedObjectStorage) (diags diag.Diagnostics) {
	if err := d.Set("configured_status", storage.ConfiguredStatus); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", storage.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	endpoints := make([]map[string]interface{}, 0)
	for _, endpoint := range storage.Endpoints {
		endpoints = append(endpoints, map[string]interface{}{
			"domain_name": endpoint.DomainName,
			"type":        endpoint.Type,
		})
	}
	if err := d.Set("endpoint", endpoints); err != nil {
		return diag.FromErr(err)
	}

	networks := make([]map[string]interface{}, 0)
	for _, network := range storage.Networks {
		networks = append(networks, map[string]interface{}{
			"family": network.Family,
			"name":   network.Name,
			"type":   network.Type,
			"uuid":   network.UUID,
		})
	}

	if err := d.Set("network", networks); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("operational_state", storage.OperationalState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("region", storage.Region); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("updated_at", storage.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	users := make([]interface{}, 0)
	for _, user := range storage.Users {
		users = append(users, user.Username)
	}
	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
