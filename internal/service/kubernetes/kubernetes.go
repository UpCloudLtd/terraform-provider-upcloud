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

const (
	clientCertificateDescription    = "TLS authentication client certificate, encoded (PEM)."
	clientKeyDescription            = "Key to pair with `client_certificate`, encoded (PEM)."
	clusterCACertificateDescription = "TLS authentication root certificate bundle, encoded (PEM)."
	hostDescription                 = "Hostname of the cluster API. Defined as URI."
	idDescription                   = "Cluster ID."
	kubeconfigDescription           = "Kubernetes config file contents for the cluster."
	nameDescription                 = "Cluster name. Needs to be unique within the account."
	networkDescription              = "Network ID for the cluster to run in."
	networkCIDRDescription          = "Network CIDR for the given network. Computed automatically."
	nodeGroupNamesDescription       = "Names of the node groups configured to cluster"
	stateDescription                = "Operational state of the cluster."
	zoneDescription                 = "Zone in which the Kubernetes cluster will be hosted, e.g. `de-fra1`."

	cleanupWaitTimeSeconds = 240
	maxResourceNameLength  = 63
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "Kubernetes cluster. NOTE: this is an experimental feature in development phase, the resource definition will change in the future.",
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      nameDescription,
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateResourceName,
			},
			"zone": {
				Description: zoneDescription,
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
				Description: nodeGroupNamesDescription,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"state": {
				Description: stateDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req := &request.CreateKubernetesClusterRequest{
		Name:    d.Get("name").(string),
		Network: d.Get("network").(string),
		Zone:    d.Get("zone").(string),
	}

	c, err := svc.CreateKubernetesCluster(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(c.UUID)

	c, err = svc.WaitForKubernetesClusterState(ctx, &request.WaitForKubernetesClusterStateRequest{
		DesiredState: upcloud.KubernetesClusterStateRunning,
		Timeout:      time.Minute * 20,
		UUID:         c.UUID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, setClusterResourceData(d, c)...)

	// No error, log a success message
	if len(diags) == 0 {
		tflog.Info(ctx, "cluster created", map[string]interface{}{"name": c.Name, "uuid": c.UUID})
	}

	return diags
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	cluster, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: d.Id()})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	return setClusterResourceData(d, cluster)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)
	if err := svc.DeleteKubernetesCluster(ctx, &request.DeleteKubernetesClusterRequest{UUID: d.Id()}); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "cluster deleted", map[string]interface{}{"name": d.Get("name").(string), "uuid": d.Id()})

	// wait before continuing so that e.g. network can be deleted (if needed)
	diags := diag.FromErr(waitForClusterToBeDeleted(ctx, svc, d.Id()))

	// If there was an error during while waiting for the cluster to be deleted - just end the delete operation here
	if len(diags) > 0 {
		return diags
	}

	// Additionally wait some time so that all cleanup operations can finish
	time.Sleep(time.Second * cleanupWaitTimeSeconds)

	return diags
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

	groups := make([]string, 0)
	for _, g := range c.NodeGroups {
		groups = append(groups, g.Name)
	}
	if err := d.Set("node_groups", groups); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func waitForClusterToBeDeleted(ctx context.Context, svc *service.Service, id string) error {
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

				return err
			}

			tflog.Info(ctx, "waiting for cluster to be deleted", map[string]interface{}{"name": c.Name, "state": c.State})
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("max retries (%d)reached while waiting for cluster to be deleted", maxRetries)
}

var validateResourceName = validation.ToDiagFunc(func(i interface{}, s string) (warns []string, errs []error) {
	val, ok := i.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("provided value is not a string"))
		return
	}

	if len(val) > maxResourceNameLength {
		errs = append(errs, fmt.Errorf("resource name (%s) too long, max allowed length is %d", val, maxResourceNameLength))
		return
	}

	nameRegexp := "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
	nameValid := regexp.MustCompile(nameRegexp).MatchString(val)
	if !nameValid {
		errs = append(errs, fmt.Errorf("name (%s) is not valid. Regular expresion used to check validation: %s", val, nameRegexp))
		return
	}

	return
})
