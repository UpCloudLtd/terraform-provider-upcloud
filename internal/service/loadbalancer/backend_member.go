package loadbalancer

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
			"backend": {
				Description: "ID of the load balancer backend to which the member is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the member must be unique within the load balancer backend service.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validator.ValidateDomainNameDiag,
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
			"backend": {
				Description: "ID of the load balancer backend to which the member is connected.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:      "The name of the member must be unique within the load balancer backend service.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validator.ValidateDomainNameDiag,
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
				Default:          0,
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
		var serviceID, beName string
		if err := utils.UnmarshalID(d.Get("backend").(string), &serviceID, &beName); err != nil {
			return diag.FromErr(err)
		}
		member, err := svc.CreateLoadBalancerBackendMember(ctx, &request.CreateLoadBalancerBackendMemberRequest{
			ServiceUUID: serviceID,
			BackendName: beName,
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

		d.SetId(utils.MarshalID(serviceID, beName, member.Name))

		tflog.Info(ctx, "backend member created", map[string]interface{}{"name": member.Name, "service_uuid": serviceID, "be_name": beName})
		return diags
	}
}

func resourceBackendMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &beName, &name); err != nil {
		return diag.FromErr(err)
	}
	member, err := svc.GetLoadBalancerBackendMember(ctx, &request.GetLoadBalancerBackendMemberRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	d.SetId(utils.MarshalID(serviceID, beName, member.Name))

	if err = d.Set("backend", utils.MarshalID(serviceID, beName)); err != nil {
		return diag.FromErr(err)
	}

	if diags = setBackendMemberResourceData(d, member); len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceBackendMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &beName, &name); err != nil {
		return diag.FromErr(err)
	}
	member, err := svc.ModifyLoadBalancerBackendMember(ctx, &request.ModifyLoadBalancerBackendMemberRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Name:        name,
		Member: request.ModifyLoadBalancerBackendMember{
			Name:        d.Get("name").(string),
			Weight:      upcloud.IntPtr(d.Get("weight").(int)),
			MaxSessions: upcloud.IntPtr(d.Get("max_sessions").(int)),
			Enabled:     upcloud.BoolPtr(d.Get("enabled").(bool)),
			IP:          upcloud.StringPtr(d.Get("ip").(string)),
			Port:        d.Get("port").(int),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, beName, member.Name))

	if diags = setBackendMemberResourceData(d, member); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "backend member updated", map[string]interface{}{"name": member.Name, "service_uuid": serviceID, "be_name": beName})
	return diags
}

func resourceBackendMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	var serviceID, beName, name string
	if err := utils.UnmarshalID(d.Id(), &serviceID, &beName, &name); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "deleting backend member", map[string]interface{}{"name": name, "service_uuid": serviceID, "be_name": beName})
	return diag.FromErr(svc.DeleteLoadBalancerBackendMember(ctx, &request.DeleteLoadBalancerBackendMemberRequest{
		ServiceUUID: serviceID,
		BackendName: beName,
		Name:        name,
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
