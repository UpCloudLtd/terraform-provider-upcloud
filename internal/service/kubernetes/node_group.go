package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
			"storage_encryption": storageEncryptionSchema("Storage encryption strategy for the nodes in this group.", true),
			"custom_plan": {
				Description: "Resource properties for custom plan",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"memory": {
							Description: "The amount of memory in megabytes to assign to individual node group node when using custom plan. Value needs to be divisible by 1024.",
							Type:        schema.TypeInt,
							ForceNew:    true,
							Required:    true,
							ValidateDiagFunc: validation.AllDiag(
								validation.ToDiagFunc(validation.IntBetween(2048, 131072)),
								validation.ToDiagFunc(validation.IntDivisibleBy(1024)),
							),
						},
						"cores": {
							Description:      "The number of CPU cores dedicated to individual node group nodes when using custom plan",
							Type:             schema.TypeInt,
							ForceNew:         true,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 20)),
						},
						"storage_size": {
							Description:      "The size of the storage device in gigabytes.",
							Type:             schema.TypeInt,
							ForceNew:         true,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(25, 1024)),
						},
						"storage_tier": {
							Description: fmt.Sprintf("The storage tier to use. Defaults to %s", upcloud.KubernetesStorageTierMaxIOPS),
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Computed:    true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
								string(upcloud.KubernetesStorageTierMaxIOPS),
								string(upcloud.KubernetesStorageTierHDD),
							}, false)),
						},
					},
				},
			},
		},
		CustomizeDiff: customdiff.Sequence(validateCustomPlan, computeClusterLevelStorageEncryption),
	}
}

func validateCustomPlan(_ context.Context, rd *schema.ResourceDiff, _ interface{}) error {
	if plan, ok := rd.Get("plan").(string); ok {
		_, customPlanOk := rd.GetOk("custom_plan")
		if !customPlanOk && plan == "custom" {
			return errors.New("`custom_plan` field is required when using custom server plan for the node group")
		}
		if customPlanOk && plan != "custom" {
			return fmt.Errorf("defining `custom_plan` properties with %s plan is not supported, use `custom` plan instead", plan)
		}
	}
	return nil
}

// computeClusterLevelStorageEncryption checks if cluster has storage encryption strategy set and applies that value to the node group when applicable.
// Purpose for this is to make storage_encryption attribute known *before* apply if it's not defined.
func computeClusterLevelStorageEncryption(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	clusterID, ok := rd.Get("cluster").(string)
	if !ok || rd.NewValueKnown("storage_encryption") {
		return nil
	}
	c, err := meta.(*service.Service).GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: clusterID})
	if err == nil && c.StorageEncryption != "" {
		return rd.SetNew("storage_encryption", c.StorageEncryption)
	}
	return nil
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

	if v, ok := d.GetOk("storage_encryption"); ok {
		req.StorageEncryption = upcloud.StorageEncryption(v.(string))
	}

	if v, ok := d.Get("custom_plan").([]interface{}); ok && len(v) > 0 {
		req.CustomPlan = &upcloud.KubernetesNodeGroupCustomPlan{
			Cores:       d.Get("custom_plan.0.cores").(int),
			Memory:      d.Get("custom_plan.0.memory").(int),
			StorageSize: d.Get("custom_plan.0.storage_size").(int),
		}
		if v, ok := d.Get("custom_plan.0.storage_tier").(string); ok && v != "" {
			req.CustomPlan.StorageTier = upcloud.StorageTier(v)
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
	d.SetId(utils.MarshalID(clusterID, ng.Name))

	ng, err = svc.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
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
	if err := utils.UnmarshalID(d.Id(), &clusterID, &name); err != nil {
		return diag.FromErr(err)
	}
	ng, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: clusterID,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}
	d.SetId(utils.MarshalID(clusterID, ng.Name))
	return setNodeGroupResourceData(d, clusterID, &ng.KubernetesNodeGroup)
}

func resourceNodeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	if !d.HasChange("node_count") {
		return nil
	}
	svc := meta.(*service.Service)
	var clusterID, name string
	if err := utils.UnmarshalID(d.Id(), &clusterID, &name); err != nil {
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
	if err := utils.UnmarshalID(d.Id(), &clusterID, &name); err != nil {
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

	if err := d.Set("storage_encryption", ng.StorageEncryption); err != nil {
		return diag.FromErr(err)
	}

	var customPlan []map[string]interface{}
	if ng.CustomPlan != nil {
		customPlan = []map[string]interface{}{
			{
				"cores":        ng.CustomPlan.Cores,
				"memory":       ng.CustomPlan.Memory,
				"storage_size": ng.CustomPlan.StorageSize,
				"storage_tier": ng.CustomPlan.StorageTier,
			},
		}
	}
	if err := d.Set("custom_plan", customPlan); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func getNodeGroupDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	c, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: id[0],
		Name:        id[1],
	})

	return map[string]interface{}{"resource": "node-group", "name": c.Name, "state": c.State}, err
}

func waitForNodeGroupToBeDeleted(ctx context.Context, svc *service.Service, clusterID, name string) error {
	return utils.WaitForResourceToBeDeleted(ctx, svc, getNodeGroupDeleted, clusterID, name)
}
