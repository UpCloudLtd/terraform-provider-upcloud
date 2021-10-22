package upcloud

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUpCloudServerBackup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUpCloudServerBackupCreate,
		ReadContext:   resourceUpCloudServerBackupRead,
		UpdateContext: resourceUpCloudServerBackupUpdate,
		DeleteContext: resourceUpCloudServerBackupDelete,
		Schema: map[string]*schema.Schema{
			"server": {
				Description: "ID of a server that should be backed up",
				Type:        schema.TypeString,
				Required:    true,
			},
			"time": {
				Description:  "Exact time at which backup should be taken in hhmm format (for example 2200)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{4}$`), "Time must be 4 digits in a hhmm format"),
			},
			"plan": {
				Description:  "Backup plan. Can be one of the following value: dailies, weeklies, monthlies",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"dailies", "weeklies", "monthlies"}, false),
			},
		},
	}
}

func setServerSimpleBackup(d *schema.ResourceData, client *service.Service) error {
	serverId := d.Get("server").(string)
	time := d.Get("time").(string)
	plan := d.Get("plan").(string)

	req := &request.ModifyServerRequest{UUID: serverId, SimpleBackup: fmt.Sprintf("%s,%s", time, plan)}
	_, err := client.ModifyServer(req)

	return err
}

func removeServerSimpleBackup(serverID string, client *service.Service) error {
	_, err := client.ModifyServer(&request.ModifyServerRequest{
		UUID:         serverID,
		SimpleBackup: "no",
	})

	return err
}

func resourceUpCloudServerBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	err := setServerSimpleBackup(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return resourceUpCloudServerBackupRead(ctx, d, meta)
}

func resourceUpCloudServerBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*service.Service)

	serverDetails, err := client.GetServerDetails(&request.GetServerDetailsRequest{UUID: d.Get("server").(string)})
	if err != nil {
		return diag.FromErr(err)
	}

	if serverDetails.SimpleBackup != "" && serverDetails.SimpleBackup != "no" {
		d.Set("server", serverDetails.UUID)

		simpleBackup := strings.Split(serverDetails.SimpleBackup, ",")
		d.Set("time", simpleBackup[0])
		d.Set("plan", simpleBackup[1])
	}

	return diags
}

func resourceUpCloudServerBackupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	o, n := d.GetChange("server")
	oldServerId := o.(string)
	newServerId := n.(string)

	if oldServerId != newServerId {
		if err := removeServerSimpleBackup(oldServerId, client); err != nil {
			return diag.FromErr(err)
		}
	}

	err := setServerSimpleBackup(d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudServerBackupRead(ctx, d, meta)
}

func resourceUpCloudServerBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*service.Service)

	serverId, _ := d.GetChange("server")
	err := removeServerSimpleBackup(serverId.(string), client)

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUpCloudServerBackupRead(ctx, d, meta)
}
