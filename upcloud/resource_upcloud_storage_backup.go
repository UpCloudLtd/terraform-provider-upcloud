package upcloud

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudStorageBackup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudStorageBackupCreate,
		ReadContext:   resourceUpCloudStorageBackupRead,
		UpdateContext: resourceUpCloudStorageBackupUpdate,
		DeleteContext: resourceUpCloudStorageBackupDelete,
		Schema: map[string]*schema.Schema{
			"storage": {
				Description: "ID of the storage to be backed up",
				Type:        schema.TypeString,
				Required:    true,
			},
			"time": {
				Description:  "Exact time at which backup should be taken in hhmm format (for example 2200)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{4}$`), "Time must be 4 digits in a hhmm format"),
			},
			"interval": {
				Description:  "Exact day at which backup should be taken",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun", "daily"}, false),
			},
			"retention": {
				Description:  "Amount of days backup should be kept for",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func setStorageBackupRule(d *schema.ResourceData, client *service.Service) error {
	storageId := d.Get("storage").(string)
	time := d.Get("time").(string)
	interval := d.Get("interval").(string)
	retention := d.Get("retention").(int)

	req := &request.ModifyStorageRequest{
		UUID: storageId,
		BackupRule: &upcloud.BackupRule{
			Time:      time,
			Interval:  interval,
			Retention: retention,
		},
	}

	if _, err := client.ModifyStorage(req); err != nil {
		return err
	}

	return nil
}

func resourceUpCloudStorageBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	err := setStorageBackupRule(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return resourceUpCloudStorageBackupRead(ctx, d, meta)
}

func resourceUpCloudStorageBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*service.Service)

	storageDetails, err := client.GetStorageDetails(&request.GetStorageDetailsRequest{UUID: d.Get("storage").(string)})
	if err != nil {
		return diag.FromErr(err)
	}

	if storageDetails.BackupRule != nil && storageDetails.BackupRule.Interval != "" {
		d.Set("storage", storageDetails.UUID)
		d.Set("time", storageDetails.BackupRule.Time)
		d.Set("interval", storageDetails.BackupRule.Interval)
		d.Set("retention", storageDetails.BackupRule.Retention)
	}

	return diags
}

func resourceUpCloudStorageBackupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	err := setStorageBackupRule(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudStorageBackupRead(ctx, d, meta)
}

func resourceUpCloudStorageBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*service.Service)

	storageId, _ := d.GetChange("storage")
	_, err := client.ModifyStorage(&request.ModifyStorageRequest{
		UUID:       storageId.(string),
		BackupRule: &upcloud.BackupRule{},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
