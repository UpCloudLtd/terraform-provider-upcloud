package loadbalancer

import (
	"context"
	"log"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceStaticBackendMember() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer's static backend member",
		CreateContext: resourceBackendMemberCreateFunc(upcloud.LoadBalancerBackendMemberTypeStatic),
		ReadContext:   resourceBackendMemberRead,
		UpdateContext: resourceBackendMemberUpdate,
		DeleteContext: resourceBackendMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"loadbalancer": {
				Description: "ID of the load balancer to which the backend is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"backend_name": {
				Description: "Name of the load balancer backend to which the member is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the member must be unique within the load balancer backend service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ip": {
				Description:      "Server IP address in the customer private network.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
			"port": {
				Description:      "Server port.",
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
			},
			"weight": {
				Description: `Used to adjust the server's weight relative to other servers. 
				All servers will receive a load proportional to their weight relative to the sum of all weights, so the higher the weight, the higher the load. 
				A value of 0 means the server will not participate in load balancing but will still accept persistent connections.`,
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
			},
			"max_sessions": {
				Description:      "Maximum number of sessions before queueing.",
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 500000)),
			},
			"enabled": {
				Description: "Indicates if the member is enabled. Disabled members are excluded from load balancing.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func ResourceDynamicBackendMember() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents load balancer's dynamic backend member",
		CreateContext: resourceBackendMemberCreateFunc(upcloud.LoadBalancerBackendMemberTypeDynamic),
		ReadContext:   resourceBackendMemberRead,
		UpdateContext: resourceBackendMemberUpdate,
		DeleteContext: resourceBackendMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"loadbalancer": {
				Description: "ID of the load balancer to which the backend is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"backend_name": {
				Description: "Name of the load balancer backend to which the member is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the member must be unique within the load balancer backend service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ip": {
				Description:      "Optional fallback IP address in case of failure on DNS resolving.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
			"port": {
				Description:      "Server port. Port is optional and can be specified in DNS SRV record.",
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
			},
			"weight": {
				Description: `Used to adjust the server's weight relative to other servers. 
				All servers will receive a load proportional to their weight relative to the sum of all weights, so the higher the weight, the higher the load. 
				A value of 0 means the server will not participate in load balancing but will still accept persistent connections.`,
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 100)),
			},
			"max_sessions": {
				Description:      "Maximum number of sessions before queueing.",
				Type:             schema.TypeInt,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 500000)),
			},
			"enabled": {
				Description: "Indicates if the member is enabled. Disabled members are excluded from load balancing.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceBackendMemberCreateFunc(memberType upcloud.LoadBalancerBackendMemberType) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
		svc := meta.(*service.Service)
		member, err := svc.CreateLoadBalancerBackendMember(&request.CreateLoadBalancerBackendMemberRequest{
			ServiceUUID: d.Get("loadbalancer").(string),
			BackendName: d.Get("backend_name").(string),
			Member: request.LoadBalancerBackendMember{
				Type:        memberType,
				Name:        d.Get("name").(string),
				Weight:      d.Get("weight").(int),
				MaxSessions: d.Get("max_sessions").(int),
				Enabled:     d.Get("enabled").(bool),
				IP:          d.Get("ip").(string),
				Port:        d.Get("port").(int),
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}

		if diags = setBackendMemberResourceData(d, member); len(diags) > 0 {
			return diags
		}

		d.SetId(member.Name)

		log.Printf("[INFO] backend member '%s' created", member.Name)
		return diags
	}
}

func resourceBackendMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	member, err := svc.GetLoadBalancerBackendMember(&request.GetLoadBalancerBackendMemberRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		BackendName: d.Get("backend_name").(string),
		Name:        d.Id(),
	})

	if err != nil {
		return handleResourceError(d.Get("name").(string), d, err)
	}

	if diags = setBackendMemberResourceData(d, member); len(diags) > 0 {
		return diags
	}

	d.SetId(member.Name)

	return diags
}

func resourceBackendMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	member, err := svc.ModifyLoadBalancerBackendMember(&request.ModifyLoadBalancerBackendMemberRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		BackendName: d.Get("backend_name").(string),
		Name:        d.Id(),
		Member: request.LoadBalancerBackendMember{
			Name:        d.Get("name").(string),
			Weight:      d.Get("weight").(int),
			MaxSessions: d.Get("max_sessions").(int),
			Enabled:     d.Get("enabled").(bool),
			IP:          d.Get("ip").(string),
			Port:        d.Get("port").(int),
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(member.Name)

	if diags = setBackendMemberResourceData(d, member); len(diags) > 0 {
		return diags
	}

	log.Printf("[INFO] backend member '%s' updated", member.Name)
	return diags
}

func resourceBackendMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	log.Printf("[INFO] deleting backend member '%s'", d.Id())
	return diag.FromErr(svc.DeleteLoadBalancerBackendMember(&request.DeleteLoadBalancerBackendMemberRequest{
		ServiceUUID: d.Get("loadbalancer").(string),
		BackendName: d.Get("backend_name").(string),
		Name:        d.Id(),
	}))
}

func setBackendMemberResourceData(d *schema.ResourceData, member *upcloud.LoadBalancerBackendMember) (diags diag.Diagnostics) {
	if err := d.Set("name", member.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("weight", member.Weight); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("max_sessions", member.MaxSessions); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("enabled", member.Enabled); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ip", member.IP); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("port", member.Port); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
