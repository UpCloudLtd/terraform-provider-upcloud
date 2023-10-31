package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceNodeGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents a node group in a Managed Kubernetes cluster.",
		CreateContext: resourceNodeGroupCreate,
		ReadContext:   resourceNodeGroupRead,
		DeleteContext: resourceNodeGroupDelete,
		UpdateContext: resourceNodeGroupUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"cluster": {
				Description: idDescription,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"node_count": {
				Description:  "Amount of nodes to provision in the node group.",
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(0),
				Required:     true,
			},
			"name": {
				Description:      "The name of the node group. Needs to be unique within a cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateResourceName,
				ForceNew:         true,
			},
			"plan": {
				Description: "The server plan used for the node group. You can list available plans with `upctl server plans`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"anti_affinity": {
				Description: `If set to true, nodes in this group will be placed on separate compute hosts.
				Please note that anti-affinity policy is considered "best effort" and enabling it does not fully guarantee that the nodes will end up on different hardware.`,
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"labels": {
				Description: "Key-value pairs to classify the node group.",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},
			"kubelet_args": {
				Description: "Additional arguments for kubelet for the nodes in this group. WARNING - those arguments will be passed directly to kubelet CLI on each worker node without any validation. Passing invalid arguments can break your whole cluster. Be extra careful when adding kubelet args.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "Kubelet argument key.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9-]+$"), ""),
							),
						},
						"value": {
							Description:      "Kubelet argument value.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 255)),
						},
					},
				},
			},
			"taint": {
				Description: "Taints for the nodes in this group.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"effect": {
							Description: "Taint effect.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
								string(upcloud.KubernetesClusterTaintEffectNoExecute),
								string(upcloud.KubernetesClusterTaintEffectNoSchedule),
								string(upcloud.KubernetesClusterTaintEffectPreferNoSchedule),
							}, false)),
						},
						"key": {
							Description: "Taint key.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9-]+$"), ""),
							),
						},
						"value": {
							Description:      "Taint value.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 255)),
						},
					},
				},
			},
			"ssh_keys": {
				Description: "You can optionally select SSH keys to be added as authorized keys to the nodes in this node group. This allows you to connect to the nodes via SSH once they are running.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"utility_network_access": {
				Description: `If set to false, nodes in this group will not have access to utility network.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
		},
	}
}

func resourceNodeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := request.KubernetesNodeGroup{
		Count:                d.Get("node_count").(int),
		Name:                 d.Get("name").(string),
		Plan:                 d.Get("plan").(string),
		AntiAffinity:         d.Get("anti_affinity").(bool),
		Labels:               []upcloud.Label{},
		SSHKeys:              []string{},
		Storage:              "",
		KubeletArgs:          []upcloud.KubernetesKubeletArg{},
		Taints:               []upcloud.KubernetesTaint{},
		UtilityNetworkAccess: upcloud.BoolPtr(d.Get("utility_network_access").(bool)),
	}
	if v, ok := d.GetOk("labels"); ok {
		for k, v := range v.(map[string]interface{}) {
			req.Labels = append(req.Labels, upcloud.Label{
				Key:   k,
				Value: v.(string),
			})
		}
	}
	if v, ok := d.GetOk("kubelet_args"); ok {
		for _, v := range v.(*schema.Set).List() {
			arg := v.(map[string]interface{})
			req.KubeletArgs = append(req.KubeletArgs, upcloud.KubernetesKubeletArg{
				Key:   arg["key"].(string),
				Value: arg["value"].(string),
			})
		}
	}
	if v, ok := d.GetOk("taint"); ok {
		for _, taint := range v.(*schema.Set).List() {
			taintData := taint.(map[string]interface{})
			effectStr := taintData["effect"].(string)

			req.Taints = append(req.Taints, upcloud.KubernetesTaint{
				Effect: upcloud.KubernetesClusterTaintEffect(effectStr),
				Key:    taintData["key"].(string),
				Value:  taintData["value"].(string),
			})
		}
	}
	if v, ok := d.GetOk("ssh_keys"); ok {
		for _, v := range v.(*schema.Set).List() {
			req.SSHKeys = append(req.SSHKeys, v.(string))
		}
	}

	clusterID := d.Get("cluster").(string)
	ng, err := svc.CreateKubernetesNodeGroup(ctx, &request.CreateKubernetesNodeGroupRequest{
		ClusterUUID: clusterID,
		NodeGroup:   req,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(marshalID(clusterID, ng.Name))

	ng, err = svc.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
		Timeout:      time.Minute * 20,
		ClusterUUID:  clusterID,
		Name:         ng.Name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setNodeGroupResourceData(d, clusterID, ng)
}

func resourceNodeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var clusterID, name string
	if err := unmarshalID(d.Id(), &clusterID, &name); err != nil {
		return diag.FromErr(err)
	}
	ng, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: clusterID,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}
	d.SetId(marshalID(clusterID, ng.Name))
	return setNodeGroupResourceData(d, clusterID, &ng.KubernetesNodeGroup)
}

func resourceNodeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	if !d.HasChange("node_count") {
		return nil
	}
	svc := meta.(*service.Service)
	var clusterID, name string
	if err := unmarshalID(d.Id(), &clusterID, &name); err != nil {
		return diag.FromErr(err)
	}
	ng, err := svc.ModifyKubernetesNodeGroup(ctx, &request.ModifyKubernetesNodeGroupRequest{
		ClusterUUID: clusterID,
		Name:        name,
		NodeGroup: request.ModifyKubernetesNodeGroup{
			Count: d.Get("node_count").(int),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ng, err = svc.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
		Timeout:      time.Minute * 20,
		ClusterUUID:  clusterID,
		Name:         ng.Name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setNodeGroupResourceData(d, clusterID, ng)
}

func resourceNodeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var clusterID, name string
	if err := unmarshalID(d.Id(), &clusterID, &name); err != nil {
		return diag.FromErr(err)
	}
	err := svc.DeleteKubernetesNodeGroup(ctx, &request.DeleteKubernetesNodeGroupRequest{
		ClusterUUID: clusterID,
		Name:        name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// wait before continuing so that all nodes are destroyed
	return diag.FromErr(waitForNodeGroupToBeDeleted(ctx, svc, clusterID, name))
}

func setNodeGroupResourceData(d *schema.ResourceData, clusterID string, ng *upcloud.KubernetesNodeGroup) (diags diag.Diagnostics) {
	if err := d.Set("cluster", clusterID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("node_count", ng.Count); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", ng.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("plan", ng.Plan); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ssh_keys", ng.SSHKeys); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("anti_affinity", ng.AntiAffinity); err != nil {
		return diag.FromErr(err)
	}

	kubeletArgs := []map[string]string{}
	for _, arg := range ng.KubeletArgs {
		kubeletArgs = append(kubeletArgs, map[string]string{
			"key":   arg.Key,
			"value": arg.Value,
		})
	}
	if err := d.Set("kubelet_args", kubeletArgs); err != nil {
		return diag.FromErr(err)
	}

	labels := map[string]string{}
	for _, lab := range ng.Labels {
		labels[lab.Key] = lab.Value
	}
	if err := d.Set("labels", labels); err != nil {
		return diag.FromErr(err)
	}

	taints := []map[string]string{}
	for _, t := range ng.Taints {
		taints = append(taints, map[string]string{
			"effect": string(t.Effect),
			"key":    t.Key,
			"value":  t.Value,
		})
	}
	if err := d.Set("taint", taints); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("utility_network_access", ng.UtilityNetworkAccess); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func waitForNodeGroupToBeDeleted(ctx context.Context, svc *service.Service, clusterID, name string) error {
	const maxRetries int = 100

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			c, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
				ClusterUUID: clusterID,
				Name:        name,
			})
			if err != nil {
				if svcErr, ok := err.(*upcloud.Problem); ok && svcErr.Status == http.StatusNotFound {
					return nil
				}

				return err
			}

			tflog.Info(ctx, "waiting for node group to be deleted", map[string]interface{}{"cluster": clusterID, "name": c.Name, "state": c.State})
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("max retries (%d)reached while waiting for node group to be deleted", maxRetries)
}
