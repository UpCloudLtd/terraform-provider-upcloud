package loadbalancer

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer service",
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "The name of the service must be unique within customer account.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNameDiagFunc,
			},
			"plan": {
				Description: "Plan which the service will have. You can list available loadbalancer plans with `upctl loadbalancer plans`",
				Type:        schema.TypeString,
				Required:    true,
			},
			"zone": {
				Description: "Zone in which the service will be hosted, e.g. `fi-hel1`. You can list available zones with `upctl zone list`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"networks": {
				ExactlyOneOf: []string{"network"},
				Description:  "Attached Networks from where traffic consumed and routed. Private networks must reside in loadbalancer zone.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     8,
				MinItems:     2,
				Elem: &schema.Resource{
					Schema: loadBalancerNetworkSchema(),
				},
			},
			"configured_status": {
				Description: "The service configured status indicates the service's current intended status. Managed by the customer.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(upcloud.LoadBalancerConfiguredStatusStarted),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{
						string(upcloud.LoadBalancerConfiguredStatusStarted),
						string(upcloud.LoadBalancerConfiguredStatusStarted),
					}, false),
				),
			},
			"frontends": {
				Description: "Frontends receive the traffic before dispatching it to the backends.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"backends": {
				Description: "Backends are groups of customer servers whose traffic should be balanced.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resolvers": {
				Description: "Domain Name Resolvers must be configured in case of customer uses dynamic type members",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"operational_state": {
				Description: "The service operational state indicates the service's current operational, effective state. Managed by the system.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"nodes": {
				Description: "Nodes are instances running load balancer service",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: loadBalancerNodeSchema(),
				},
			},
			"network": {
				Deprecated:  "Use 'networks' to define networks attached to load balancer",
				Description: "Private network UUID where traffic will be routed. Must reside in load balancer zone.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
			},
			"dns_name": {
				Deprecated:  "Use 'networks' to get network DNS name",
				Description: "DNS name of the load balancer",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"labels": utils.LabelsSchema("load balancer"),
		},
		CustomizeDiff: customdiff.ForceNewIfChange("networks.#", func(ctx context.Context, old, new, meta interface{}) bool {
			return new.(int) != old.(int)
		}),
	}
}

func loadBalancerNetworkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "The name of the network must be unique within the service.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.All(
				validation.StringLenBetween(0, 65),
				validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]+$"), ""),
			)),
		},
		"type": {
			Description: "The type of the network. Only one public network can be attached and at least one private network must be attached.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{
					string(upcloud.LoadBalancerNetworkTypePrivate),
					string(upcloud.LoadBalancerNetworkTypePublic),
				}, false),
			),
		},
		"family": {
			Description: "Network family. Currently only `IPv4` is supported.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{
					string(upcloud.LoadBalancerAddressFamilyIPv4),
				}, false),
			),
		},
		"network": {
			Description: "Private network UUID. Required for private networks and must reside in loadbalancer zone. For public network the field should be omitted.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
		},
		"dns_name": {
			Description: "DNS name of the load balancer network",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"id": {
			Description: "Network identifier.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func loadBalancerNodeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"operational_state": {
			Description: "Node's operational state. Managed by the system.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"networks": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: loadBalancerNodeNetworkSchema(),
			},
		},
	}
}

func loadBalancerNodeNetworkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "The name of the network.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: "The type of the network.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"ip_addresses": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"address": {
						Description: "Node's IP address.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"listen": {
						Description: "Does IP address listen network connections.",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	networks, err := loadBalancerNetworksFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}
	req := &request.CreateLoadBalancerRequest{
		Name:             d.Get("name").(string),
		Plan:             d.Get("plan").(string),
		Zone:             d.Get("zone").(string),
		NetworkUUID:      d.Get("network").(string),
		ConfiguredStatus: upcloud.LoadBalancerConfiguredStatus(d.Get("configured_status").(string)),
		Frontends:        []request.LoadBalancerFrontend{},
		Backends:         []request.LoadBalancerBackend{},
		Resolvers:        []request.LoadBalancerResolver{},
		Networks:         networks,
		Labels:           utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{})),
	}
	lb, err := svc.CreateLoadBalancer(ctx, req)
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(lb.UUID)

	if diags = setLoadBalancerResourceData(d, lb); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "load balancer created", map[string]interface{}{"name": lb.Name, "uuid": lb.UUID})
	return diags
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var err error
	svc := meta.(*service.Service)
	lb, err := svc.GetLoadBalancer(ctx, &request.GetLoadBalancerRequest{UUID: d.Id()})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setLoadBalancerResourceData(d, lb); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	// handle network renaming before modifying load balancer so that new network names are present in `lb` before setting state
	if diags = resourceLoadBalancerNetworkUpdate(ctx, d, svc); len(diags) > 0 {
		return diags
	}

	req := &request.ModifyLoadBalancerRequest{
		UUID:             d.Id(),
		Name:             d.Get("name").(string),
		Plan:             d.Get("plan").(string),
		ConfiguredStatus: d.Get("configured_status").(string),
	}

	if d.HasChange("labels") {
		labels := utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{}))
		req.Labels = &labels
	}

	lb, err := svc.ModifyLoadBalancer(ctx, req)
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setLoadBalancerResourceData(d, lb); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "load balancer updated", map[string]interface{}{"name": lb.Name, "uuid": lb.UUID})
	return diags
}

func resourceLoadBalancerNetworkUpdate(ctx context.Context, d *schema.ResourceData, svc *service.Service) (diags diag.Diagnostics) {
	if !d.HasChange("networks") {
		return nil
	}
	if nets, ok := d.GetOk("networks"); ok {
		for i := range nets.([]interface{}) {
			key := fmt.Sprintf("networks.%d.name", i)
			if d.HasChange(key) {
				if name, ok := d.Get(key).(string); ok {
					var serviceID, networkName, id string
					idKey := fmt.Sprintf("networks.%d.id", i)
					if id, ok = d.Get(idKey).(string); !ok {
						return diag.FromErr(fmt.Errorf("unable to determine network ID %s", idKey))
					}
					if err := utils.UnmarshalID(id, &serviceID, &networkName); err != nil {
						return diag.FromErr(err)
					}
					req := &request.ModifyLoadBalancerNetworkRequest{
						ServiceUUID: serviceID,
						Name:        networkName,
						Network:     request.ModifyLoadBalancerNetwork{Name: name},
					}
					if _, err := svc.ModifyLoadBalancerNetwork(ctx, req); err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}
	return diags
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)
	if err := svc.DeleteLoadBalancer(ctx, &request.DeleteLoadBalancerRequest{UUID: d.Id()}); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "load balancer deleted", map[string]interface{}{"name": d.Get("name").(string), "uuid": d.Id()})

	// Wait load balancer to shutdown before continuing so that e.g. network can be deleted (if needed)
	return diag.FromErr(waitLoadBalancerToShutdown(ctx, svc, d.Id()))
}

func setLoadBalancerResourceData(d *schema.ResourceData, lb *upcloud.LoadBalancer) (diags diag.Diagnostics) {
	if err := d.Set("name", lb.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("plan", lb.Plan); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", lb.Zone); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configured_status", lb.ConfiguredStatus); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("operational_state", lb.OperationalState); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("dns_name", lb.DNSName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("labels", utils.LabelsSliceToMap(lb.Labels)); err != nil {
		return diag.FromErr(err)
	}

	var frontends, backends, resolvers []string

	for _, f := range lb.Frontends {
		frontends = append(frontends, f.Name)
	}

	if err := d.Set("frontends", frontends); err != nil {
		return diag.FromErr(err)
	}

	for _, b := range lb.Backends {
		backends = append(backends, b.Name)
	}

	if err := d.Set("backends", backends); err != nil {
		return diag.FromErr(err)
	}

	for _, r := range lb.Resolvers {
		resolvers = append(resolvers, r.Name)
	}

	if err := d.Set("resolvers", resolvers); err != nil {
		return diag.FromErr(err)
	}

	nodes := make([]map[string]interface{}, 0)
	for _, n := range lb.Nodes {
		node := map[string]interface{}{
			"operational_state": n.OperationalState,
		}
		networks := make([]map[string]interface{}, 0)
		for _, net := range n.Networks {
			ips := make([]map[string]interface{}, 0)
			for _, ip := range net.IPAddresses {
				ips = append(ips, map[string]interface{}{
					"address": ip.Address,
					"listen":  ip.Listen,
				})
			}
			networks = append(networks, map[string]interface{}{
				"name":         net.Name,
				"type":         net.Type,
				"ip_addresses": ips,
			})
		}
		node["networks"] = networks
		nodes = append(nodes, node)
	}
	if err := d.Set("nodes", nodes); err != nil {
		return diag.FromErr(err)
	}
	return setLoadBalancerNetworkResourceData(d, lb)
}

func setLoadBalancerNetworkResourceData(d *schema.ResourceData, lb *upcloud.LoadBalancer) (diags diag.Diagnostics) {
	// If legacy network UUID is set do not populate state with autogenerated network objects
	if lb.NetworkUUID != "" {
		if err := d.Set("networks", nil); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("network", lb.NetworkUUID); err != nil {
			return diag.FromErr(err)
		}
		return
	}

	// network objects are in use, reset network field
	if err := d.Set("network", nil); err != nil {
		return diag.FromErr(err)
	}
	networks := make([]map[string]string, 0)
	for _, net := range lb.Networks {
		networks = append(networks, map[string]string{
			"name":     net.Name,
			"type":     string(net.Type),
			"family":   string(net.Family),
			"network":  net.UUID,
			"dns_name": net.DNSName,
			"id":       utils.MarshalID(lb.UUID, net.Name),
		})
	}
	if err := d.Set("networks", networks); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func waitLoadBalancerToShutdown(ctx context.Context, svc *service.Service, id string) error {
	const maxRetries int = 100
	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			lb, err := svc.GetLoadBalancer(ctx, &request.GetLoadBalancerRequest{UUID: id})
			if err != nil {
				if svcErr, ok := err.(*upcloud.Problem); ok && svcErr.Status == http.StatusNotFound {
					return nil
				}
				return err
			}
			tflog.Info(ctx, "waiting load balancer to shutdown", map[string]interface{}{"name": lb.Name, "state": lb.OperationalState})
		}
		time.Sleep(5 * time.Second)
	}
	return errors.New("max retries reached while waiting for load balancer instance to shutdown")
}

func loadBalancerNetworksFromResourceData(d *schema.ResourceData) ([]request.LoadBalancerNetwork, error) {
	req := make([]request.LoadBalancerNetwork, 0)
	if nets, ok := d.GetOk("networks"); ok {
		for i, n := range nets.([]interface{}) {
			n := n.(map[string]interface{})
			r := request.LoadBalancerNetwork{
				Name:   n["name"].(string),
				Type:   upcloud.LoadBalancerNetworkType(n["type"].(string)),
				Family: upcloud.LoadBalancerAddressFamily(n["family"].(string)),
				UUID:   n["network"].(string),
			}
			if r.Type == upcloud.LoadBalancerNetworkTypePrivate && r.UUID == "" {
				return req, fmt.Errorf("load balancer's private network (#%d) ID is required", i)
			}
			if r.Type == upcloud.LoadBalancerNetworkTypePublic && r.UUID != "" {
				return req, fmt.Errorf("setting load balancer's public network (#%d) ID is not supported", i)
			}
			req = append(req, r)
		}
	}
	return req, nil
}
