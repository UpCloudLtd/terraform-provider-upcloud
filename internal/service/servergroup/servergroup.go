package servergroup

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const serverGroupNotFoundErrorCode = "GROUP_NOT_FOUND"

func ResourceServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerGroupCreate,
		ReadContext:   resourceServerGroupRead,
		UpdateContext: resourceServerGroupUpdate,
		DeleteContext: resourceServerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Description: "Title of your server group",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"labels": {
				Description: "Labels for your server group",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"members": {
				Description: "UUIDs of the servers that are members of this group",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"anti_affinity": {
				Description: `Is group an anti-affinity group. Setting this to true will result in
				all servers in the group being placed on separate compute hosts.

				NOTE: this is an experimental feature. The anti-affinity policy is "best-effort" and it is not
				guaranteed that all the servers will end up on a separate compute hosts. You can verify if the
				anti-affinity policies are met by requesting a server group details from API. For more information
				please see UpCloud API documentation on server groups
				
				NOTE: anti-affinity policies are only applied on server start. This means that if anti-affinity
				policies in server group are not met, you need to manually restart the servers in said group,
				for example via API, UpCloud Control Panel or upctl (UpCloud CLI)`,
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	baseErrMsg := "creating server group failed"

	req, err := createServerGroupRequestFromConfig(ctx, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  baseErrMsg,
			Detail:   err.Error(),
		})

		return diags
	}

	group, err := svc.CreateServerGroup(ctx, req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  baseErrMsg,
			Detail:   err.Error(),
		})

		return diags
	}

	d.SetId(group.UUID)

	diags = append(diags, resourceServerGroupRead(ctx, d, meta)...)
	return diags
}

func resourceServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	baseErrMsg := "reading server group data failed"

	group, err := svc.GetServerGroup(ctx, &request.GetServerGroupRequest{UUID: d.Id()})
	if err != nil {
		if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == serverGroupNotFoundErrorCode {
			diags = append(diags, utils.DiagBindingRemovedWarningFromUpcloudErr(svcErr, d.Get("name").(string)))
			d.SetId("")
			return diags
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  baseErrMsg,
			Detail:   err.Error(),
		})

		return diags
	}

	err = setServerGroupData(group, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  baseErrMsg,
			Detail:   err.Error(),
		})
	}

	return diags
}

func resourceServerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	baseErrMsg := "modifying server group data failed"

	req, err := modifyServerGroupRequestFromConfig(ctx, d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  baseErrMsg,
			Detail:   err.Error(),
		})
		return diags
	}

	_, err = svc.ModifyServerGroup(ctx, req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  baseErrMsg,
			Detail:   err.Error(),
		})
		return diags
	}

	diags = append(diags, resourceServerGroupRead(ctx, d, meta)...)
	return diags
}

func resourceServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	err := svc.DeleteServerGroup(ctx, &request.DeleteServerGroupRequest{UUID: d.Id()})
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func setServerGroupData(group *upcloud.ServerGroup, d *schema.ResourceData) error {
	if group.Title != "" {
		if err := d.Set("title", group.Title); err != nil {
			return err
		}
	}

	if err := d.Set("anti_affinity", group.AntiAffinity.Bool()); err != nil {
		return err
	}

	if err := d.Set("members", group.Members); err != nil {
		return err
	}

	if err := d.Set("labels", utils.LabelSliceToMap(group.Labels)); err != nil {
		return err
	}

	return nil
}

func createServerGroupRequestFromConfig(ctx context.Context, d *schema.ResourceData) (*request.CreateServerGroupRequest, error) {
	result := &request.CreateServerGroupRequest{
		Title: d.Get("title").(string),
	}

	if d.Get("anti_affinity").(bool) {
		result.AntiAffinity = upcloud.True
	} else {
		result.AntiAffinity = upcloud.False
	}

	members, membersWereSet := d.GetOk("members")
	if membersWereSet {
		membersSlice, err := utils.SetOfStringsToSlice(ctx, members)
		if err != nil {
			return result, err
		}

		result.Members = membersSlice
	}

	labels, labelsWereSet := d.GetOk("labels")
	if labelsWereSet {
		labelsSlice, err := utils.MapOfStringsToLabelSlice(ctx, labels)
		if err != nil {
			return result, err
		}

		result.Labels = &labelsSlice
	}

	return result, nil
}

func modifyServerGroupRequestFromConfig(ctx context.Context, d *schema.ResourceData) (*request.ModifyServerGroupRequest, error) {
	result := &request.ModifyServerGroupRequest{
		Title: d.Get("title").(string),
		UUID:  d.Id(),
	}

	if d.Get("anti_affinity").(bool) {
		result.AntiAffinity = upcloud.True
	} else {
		result.AntiAffinity = upcloud.False
	}

	if d.HasChange("members") {
		members, err := utils.SetOfStringsToSlice(ctx, d.Get("members"))
		if err != nil {
			return result, err
		}

		membersUUIDSlice := utils.SliceOfStringToServerUUIDSlice(members)
		result.Members = &membersUUIDSlice
	}

	if d.HasChange("labels") {
		labels, err := utils.MapOfStringsToLabelSlice(ctx, d.Get("labels"))
		if err != nil {
			return result, err
		}

		result.Labels = &labels
	}

	return result, nil
}
