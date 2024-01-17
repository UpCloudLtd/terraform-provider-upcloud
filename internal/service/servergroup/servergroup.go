package servergroup

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	titleDescription        = "Title of your server group"
	membersDescription      = "UUIDs of the servers that are members of this group" // TODO(#469): add warning about server.server_group
	trackMembersDescription = "Controls if members of the server group are being tracked in this resource. Set to `false` when using `server_group` property of `upcloud_server` to attach servers to the server group to avoid delayed state updates."
	// Lines > 1 should have one level of indentation to keep them under the right list item
	antiAffinityPolicyDescription = `Defines if a server group is an anti-affinity group. Setting this to ` + "`strict` or `yes`" + ` will
	result in all servers in the group being placed on separate compute hosts. The value can be ` + "`strict`, `yes`, or `no`" + `.

	* ` + "`strict`" + ` policy doesn't allow servers in the same server group to be on the same host
	* ` + "`yes`" + ` refers to best-effort policy and tries to put servers on different hosts, but this is not guaranteed
	* ` + "`no`" + ` refers to having no policy and thus no effect on server host affinity

	To verify if the anti-affinity policies are met by requesting a server group details from API. For more information
	please see UpCloud API documentation on server groups.

	Plese also note that anti-affinity policies are only applied on server start. This means that if anti-affinity
	policies in server group are not met, you need to manually restart the servers in said group,
	for example via API, UpCloud Control Panel or upctl (UpCloud CLI)`
)

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
				Description: titleDescription,
				Type:        schema.TypeString,
				Required:    true,
			},
			"labels": utils.LabelsSchema("server group"),
			"members": {
				Description: membersDescription,
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"track_members": {
				Description: trackMembersDescription,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				// TODO CustomizeDiff to validate when track_members == false members must be empty
			},
			"anti_affinity_policy": {
				Description: antiAffinityPolicyDescription,
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "no",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					string(upcloud.ServerGroupAntiAffinityPolicyBestEffort),
					string(upcloud.ServerGroupAntiAffinityPolicyOff),
					string(upcloud.ServerGroupAntiAffinityPolicyStrict),
				}, false)),
			},
		},
	}
}

func resourceServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
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
	svc := meta.(*service.Service)
	baseErrMsg := "reading server group data failed"

	group, err := svc.GetServerGroup(ctx, &request.GetServerGroupRequest{UUID: d.Id()})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
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
	svc := meta.(*service.Service)
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
	svc := meta.(*service.Service)

	err := svc.DeleteServerGroup(ctx, &request.DeleteServerGroupRequest{UUID: d.Id()})
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func setServerGroupData(group *upcloud.ServerGroup, d *schema.ResourceData) error {
	if err := d.Set("title", group.Title); err != nil {
		return err
	}

	if err := d.Set("anti_affinity_policy", group.AntiAffinityPolicy); err != nil {
		return err
	}

	if d.Get("track_members").(bool) {
		if err := d.Set("members", group.Members); err != nil {
			return err
		}
	}

	return d.Set("labels", utils.LabelSliceToMap(group.Labels))
}

func createServerGroupRequestFromConfig(ctx context.Context, d *schema.ResourceData) (*request.CreateServerGroupRequest, error) {
	result := &request.CreateServerGroupRequest{
		Title: d.Get("title").(string),
	}

	aaPolicy := d.Get("anti_affinity_policy").(string)
	result.AntiAffinityPolicy = upcloud.ServerGroupAntiAffinityPolicy(aaPolicy)

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

	aaPolicy := d.Get("anti_affinity_policy").(string)
	result.AntiAffinityPolicy = upcloud.ServerGroupAntiAffinityPolicy(aaPolicy)

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
