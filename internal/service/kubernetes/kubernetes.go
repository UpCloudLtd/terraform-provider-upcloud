package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	clientCertificateDescription     = "TLS authentication client certificate, encoded (PEM)."
	clientKeyDescription             = "Key to pair with `client_certificate`, encoded (PEM)."
	clusterCACertificateDescription  = "TLS authentication root certificate bundle, encoded (PEM)."
	hostDescription                  = "Hostname of the cluster API. Defined as URI."
	idDescription                    = "Cluster ID."
	kubeconfigDescription            = "Kubernetes config file contents for the cluster."
	nameDescription                  = "Cluster name. Needs to be unique within the account."
	networkDescription               = "Network ID for the cluster to run in."
	networkCIDRDescription           = "Network CIDR for the given network. Computed automatically."
	nodeGroupsCountDescription       = "Amount of nodes to provision in the node group."
	nodeGroupsDescription            = "Node groups for workloads. Currently not available in state, although created."
	nodeGroupsLabelsDescription      = "Key-value pairs to classify the node group."
	nodeGroupsNameDescription        = "The name of the node group. Needs to be unique within a cluster."
	nodeGroupsPlanDescription        = "The pricing plan used for the node group. Valid values available via `upcloud_kubernetes_plan` datasource field `description`."
	nodeGroupsSSHKeysDescription     = "You can optionally select SSH keys to be added as authorized keys to the nodes in this node group. This allows you to connect to the nodes via SSH once they are running."
	nodeGroupsKubeletArgsDescription = "Additional arguments for kubelet for the nodes in this group."
	nodeGroupTaintsDescription       = "Taints for the nodes in this group."
	nodeGroupTaintEffectDescription  = "Taint effect."
	nodeGroupTaintKeyDescription     = "Taint key."
	nodeGroupTaintValueDescription   = "Taint value."
	planNameDescription              = "The name used to identify a pricing plan, e.g. `large`."
	planDescriptionDescription       = "The description of a pricing plan. e.g. `K8S-2xCPU-4GB`."
	stateDescription                 = "Operational state of the cluster."
	storageDescription               = "Storage template ID for node groups."
	typeDescription                  = "Cluster type. Values: `standalone`"
	zoneDescription                  = "Zone in which the Kubernetes cluster will be hosted, e.g. `de-fra1`."
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "Kubernetes cluster",
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: nameDescription,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"network": {
				Description: networkDescription,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"network_cidr": {
				Description: networkCIDRDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"node_groups": {
				Description: nodeGroupsDescription,
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Description: nodeGroupsCountDescription,
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
						},
						"labels": {
							Description: nodeGroupsLabelsDescription,
							Type:        schema.TypeMap,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"kubelet_args": {
							Description: nodeGroupsKubeletArgsDescription,
							Type:        schema.TypeMap,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						"taint": {
							Description: nodeGroupTaintsDescription,
							Type:        schema.TypeSet,
							Optional:    true,
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
										Description: "Taint key",
										Type:        schema.TypeString,
										Required:    true,
									},
									"value": {
										Description: "Taint value",
										Type:        schema.TypeString,
										Required:    true,
									},
								},
							},
						},
						"name": {
							Description: nodeGroupsNameDescription,
							Type:        schema.TypeString,
							Required:    true,
						},
						"plan": {
							Description: nodeGroupsPlanDescription,
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"ssh_keys": {
							Description: nodeGroupsSSHKeysDescription,
							Type:        schema.TypeSet,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"state": {
				Description: stateDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"zone": {
				Description: zoneDescription,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)

	req := &request.CreateKubernetesClusterRequest{
		Name:       d.Get("name").(string),
		Network:    d.Get("network").(string),
		NodeGroups: getNodeGroupsFromConfig(d),
		Zone:       d.Get("zone").(string),
	}

	c, err := svc.CreateKubernetesCluster(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(c.UUID)

	c, err = svc.WaitForKubernetesClusterState(ctx, &request.WaitForKubernetesClusterStateRequest{
		DesiredState: upcloud.KubernetesClusterStateRunning,
		Timeout:      time.Minute * 5,
		UUID:         c.UUID,
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "cluster did not reach ready state",
			Detail:   err.Error(),
		})
	}

	diags = append(diags, setClusterResourceData(d, c)...)

	// No error, log a success message
	if len(diags) == 0 {
		tflog.Info(ctx, "cluster created", map[string]interface{}{"name": c.Name, "uuid": c.UUID})
	}

	return diags
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.ServiceContext)
	cluster, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: d.Id()})
	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	return setClusterResourceData(d, cluster)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.ServiceContext)
	if err := svc.DeleteKubernetesCluster(ctx, &request.DeleteKubernetesClusterRequest{UUID: d.Id()}); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "cluster deleted", map[string]interface{}{"name": d.Get("name").(string), "uuid": d.Id()})

	// wait before continuing so that e.g. network can be deleted (if needed)
	return diag.FromErr(waitForClusterToBeDeleted(ctx, svc, d.Id()))
}

func setClusterResourceData(d *schema.ResourceData, c *upcloud.KubernetesCluster) (diags diag.Diagnostics) {
	if err := d.Set("name", c.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("network", c.Network); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("network_cidr", c.NetworkCIDR); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("state", c.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", c.Zone); err != nil {
		return diag.FromErr(err)
	}

	if err := setClusterNodeGroupsData(d, c); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setClusterNodeGroupsData(d *schema.ResourceData, c *upcloud.KubernetesCluster) error {
	result := []map[string]interface{}{}

	for _, g := range c.NodeGroups {
		group := map[string]interface{}{}
		group["count"] = g.Count
		group["name"] = g.Name
		group["plan"] = g.Plan
		group["ssh_keys"] = g.SSHKeys

		kubeletArgs := map[string]string{}
		for _, arg := range g.KubeletArgs {
			kubeletArgs[arg.Key] = arg.Value
		}
		group["kubelet_args"] = kubeletArgs

		labels := map[string]string{}
		for _, lab := range g.Labels {
			labels[lab.Key] = lab.Value
		}
		group["labels"] = labels

		taints := []map[string]string{}
		for _, t := range g.Taints {
			taints = append(taints, map[string]string{
				"effect": string(t.Effect),
				"key":    t.Key,
				"value":  t.Value,
			})
		}
		group["taint"] = taints

		result = append(result, group)
	}

	return d.Set("node_groups", result)
}

func waitForClusterToBeDeleted(ctx context.Context, svc *service.ServiceContext, id string) error {
	const maxRetries int = 100

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			c, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: id})
			if err != nil {
				if svcErr, ok := err.(*upcloud.Problem); ok && svcErr.Status == http.StatusNotFound {
					return nil
				}

				// Support for legacy-style API errors
				// TODO: remove when all API endpoints support the json+problem error handling
				if svcErr, ok := err.(*upcloud.Error); ok && svcErr.ErrorCode == "NotFound" {
					return nil
				}

				return err
			}

			tflog.Info(ctx, "waiting for cluster to be deleted", map[string]interface{}{"name": c.Name, "state": c.State})
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("max retries (%d)reached while waiting for cluster to be deleted", maxRetries)
}

func getNodeGroupsFromConfig(d *schema.ResourceData) []upcloud.KubernetesNodeGroup {
	result := make([]upcloud.KubernetesNodeGroup, 0)
	config := d.Get("node_groups").(*schema.Set)

	for _, el := range config.List() {
		data := el.(map[string]interface{})

		sshKeys := []string{}
		for _, key := range data["ssh_keys"].(*schema.Set).List() {
			sshKeys = append(sshKeys, key.(string))
		}

		labels := upcloud.LabelSlice{}
		for k, v := range data["labels"].(map[string]interface{}) {
			labels = append(labels, upcloud.Label{
				Key:   k,
				Value: v.(string),
			})
		}

		kubeletArgs := []upcloud.KubernetesKubeletArg{}
		for k, v := range data["kubelet_args"].(map[string]interface{}) {
			kubeletArgs = append(kubeletArgs, upcloud.KubernetesKubeletArg{
				Key:   k,
				Value: v.(string),
			})
		}

		taints := []upcloud.KubernetesTaint{}
		for _, taint := range data["taint"].(*schema.Set).List() {
			taintData := taint.(map[string]interface{})
			effectStr := taintData["effect"].(string)

			taints = append(taints, upcloud.KubernetesTaint{
				Effect: upcloud.KubernetesClusterTaintEffect(effectStr),
				Key:    taintData["key"].(string),
				Value:  taintData["value"].(string),
			})
		}

		result = append(result, upcloud.KubernetesNodeGroup{
			Count:       data["count"].(int),
			Name:        data["name"].(string),
			Plan:        data["plan"].(string),
			SSHKeys:     sshKeys,
			Labels:      labels,
			KubeletArgs: kubeletArgs,
			Taints:      taints,
		})
	}

	return result
}
