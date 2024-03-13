package gateway

import (
	"context"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "The name of the connection, should be unique within the gateway",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateName,
			},
			"gateway": {
				Description: "The UUID of the gateway to which the connection belongs",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description: "The type of the connection; currently the only supported type is 'ipsec'",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "ipsec",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{"ipsec"}, false),
				),
			},
			"local_route": {
				Description:  "Route for the UpCloud side of the network",
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"local_route", "remote_route"},
				Elem:         gatewayRouteSchema(),
			},
			"remote_route": {
				Description:  "Route for the remote side of the network",
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"local_route", "remote_route"},
				Elem:         gatewayRouteSchema(),
			},
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	serviceID := d.Get("gateway").(string)

	conn, err := svc.CreateGatewayConnection(ctx, &request.CreateGatewayConnectionRequest{
		ServiceUUID: serviceID,
		Connection: request.GatewayConnection{
			Name:         d.Get("name").(string),
			Type:         upcloud.GatewayConnectionType(d.Get("type").(string)),
			LocalRoutes:  expandRoutes(d.Get("local_route")),
			RemoteRoutes: expandRoutes(d.Get("remote_route")),
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceID, conn.Name))

	diags = append(diags, setConnectionResourceData(d, conn)...)
	if len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "gateway connection created", map[string]interface{}{"name": conn.Name, "service_uuid": serviceID})
	return diags
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc         = meta.(*service.Service)
		serviceUUID string
		name        string
	)

	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &name); err != nil {
		return diag.FromErr(err)
	}

	conn, err := svc.GetGatewayConnection(ctx, &request.GetGatewayConnectionRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	})
	if err != nil {
		return utils.HandleResourceError(name, d, err)
	}

	d.SetId(utils.MarshalID(serviceUUID, conn.Name))

	if err = d.Set("gateway", serviceUUID); err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, setConnectionResourceData(d, conn)...)
	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc         = meta.(*service.Service)
		serviceUUID string
		name        string
	)

	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &name); err != nil {
		return diag.FromErr(err)
	}

	conn, err := svc.ModifyGatewayConnection(ctx, &request.ModifyGatewayConnectionRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
		Connection: request.ModifyGatewayConnection{
			LocalRoutes:  expandRoutes(d.Get("local_route")),
			RemoteRoutes: expandRoutes(d.Get("remote_route")),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, setConnectionResourceData(d, conn)...)
	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc         = meta.(*service.Service)
		serviceUUID string
		name        string
	)

	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &name); err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "deleting gateway connection", map[string]interface{}{"name": name, "service_uuid": serviceUUID})

	return diag.FromErr(svc.DeleteGatewayConnection(ctx, &request.DeleteGatewayConnectionRequest{
		ServiceUUID: serviceUUID,
		Name:        name,
	}))
}

func gatewayRouteSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "Type of route; currently the only supported type is 'static'",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "static",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{"static"}, false),
				),
			},
			"static_network": {
				Description:      "Destination prefix of the route; needs to be a valid IPv4 prefix",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
			},
			"name": {
				Description:      "Name of the route",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateName,
			},
		},
	}
}

func setConnectionResourceData(d *schema.ResourceData, conn *upcloud.GatewayConnection) (diags diag.Diagnostics) {
	if err := d.Set("name", conn.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("type", conn.Type); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("local_route", flattenRoutes(conn.LocalRoutes)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("remote_route", flattenRoutes(conn.RemoteRoutes)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenRoutes(routes []upcloud.GatewayRoute) []interface{} {
	var data = make([]interface{}, len(routes))
	for i, route := range routes {
		data[i] = map[string]interface{}{
			"type":           route.Type,
			"static_network": route.StaticNetwork,
			"name":           route.Name,
		}
	}
	return data
}

func expandRoutes(d interface{}) []upcloud.GatewayRoute {
	routes := d.(*schema.Set).List()
	var data = make([]upcloud.GatewayRoute, len(routes))
	for i, route := range routes {
		route := route.(map[string]interface{})
		data[i] = upcloud.GatewayRoute{
			Type:          upcloud.GatewayRouteType(route["type"].(string)),
			StaticNetwork: route["static_network"].(string),
			Name:          route["name"].(string),
		}
	}
	return data
}
