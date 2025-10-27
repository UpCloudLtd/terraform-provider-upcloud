package gateway

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	nameDescription             = "Gateway name. Needs to be unique within the account."
	zoneDescription             = "Zone in which the gateway will be hosted, e.g. `de-fra1`."
	featuresDescription         = "Features enabled for the gateway. Not all VPN plans allow VPN feature., for more details see https://upcloud.com/resources/docs/networking#nat-and-vpn-gateways."
	routerDescription           = "Attached Router from where traffic is routed towards the network gateway service."
	routerIDDescription         = "ID of the router attached to the gateway."
	configuredStatusDescription = "The service configured status indicates the service's current intended status. Managed by the customer."
	operationalStateDescription = "The service operational state indicates the service's current operational, effective state. Managed by the system."
	addressesDescription        = "IP addresses assigned to the gateway."
	planDescription             = "Gateway pricing plan."
	connectionsDescription      = "Names of connections attached to the gateway. Note that this field can have outdated information as connections are created by a separate resource. To make sure that you have the most recent data run 'terrafrom refresh'."

	cleanupWaitTimeSeconds = 15
)

func ResourceGateway() *schema.Resource {
	return &schema.Resource{
		Description:   "Network gateways connect SDN Private Networks to external IP networks.",
		CreateContext: resourceGatewayCreate,
		ReadContext:   resourceGatewayRead,
		UpdateContext: resourceGatewayUpdate,
		DeleteContext: resourceGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      nameDescription,
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateName,
			},
			"zone": {
				Description: zoneDescription,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"features": {
				Description: featuresDescription,
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validateFeaturesElen,
				},
			},
			"router": {
				Description: routerDescription,
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MaxItems:    1,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: routerIDDescription,
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"labels": utils.LabelsSchema("network gateway"),
			"configured_status": {
				Description:      configuredStatusDescription,
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(upcloud.GatewayConfiguredStatusStarted),
				ValidateDiagFunc: validateConfiguredStatus,
			},
			"operational_state": {
				Description: operationalStateDescription,
				Type:        schema.TypeString,
				Computed:    true,
			},
			"plan": {
				Description: planDescription,
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"address": {
				Description: addressesDescription,
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:        schema.TypeString,
							Description: "IP addresss",
							Computed:    true,
							Optional:    false,
							Required:    false,
						},
						"name": {
							Type:             schema.TypeString,
							Description:      "Name of the IP address",
							Computed:         true,
							Optional:         true,
							ValidateDiagFunc: validateName,
						},
					},
				},
			},
			"connections": {
				Description: connectionsDescription,
				Computed:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"addresses": {
				Deprecated:  "Use 'address' attribute instead. This attribute will be removed in the next major version of the provider",
				Description: addressesDescription,
				Computed:    true,
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:        schema.TypeString,
							Description: "IP addresss",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the IP address",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	features := []upcloud.GatewayFeature{}
	for _, i := range d.Get("features").(*schema.Set).List() {
		features = append(features, upcloud.GatewayFeature(i.(string)))
	}

	req := &request.CreateGatewayRequest{
		Name:     d.Get("name").(string),
		Zone:     d.Get("zone").(string),
		Plan:     d.Get("plan").(string),
		Features: features,
		Routers: []request.GatewayRouter{
			{UUID: d.Get("router.0.id").(string)},
		},
		Labels:           utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{})),
		ConfiguredStatus: upcloud.GatewayConfiguredStatus(d.Get("configured_status").(string)),
	}

	addresses := []upcloud.GatewayAddress{}
	for i, val := range d.Get("address").(*schema.Set).List() {
		addrMap := val.(map[string]interface{})

		addrName, ok := addrMap["name"]
		if !ok || addrName == "" {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Malformed resource data",
				Detail:        "Gateway address does not have a required name.",
				AttributePath: cty.GetAttrPath("address").IndexInt(i).GetAttr("name"),
			})
		}

		addresses = append(addresses, upcloud.GatewayAddress{
			Name: addrName.(string),
		})
	}

	if len(addresses) > 0 {
		req.Addresses = addresses
	}

	gw, err := svc.CreateGateway(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(gw.UUID)

	gw, err = waitForGatewayToBeRunning(ctx, svc, gw.UUID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, setGatewayResourceData(d, gw)...)

	// No error, log a success message
	if len(diags) == 0 {
		tflog.Info(ctx, "network gateway created", map[string]interface{}{"name": gw.Name, "uuid": gw.UUID})
	}

	return diags
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)
	gw, err := svc.GetGateway(ctx, &request.GetGatewayRequest{UUID: d.Id()})
	if err != nil {
		return utils.HandleResourceError(d.Get("name").(string), d, err)
	}

	return setGatewayResourceData(d, gw)
}

func resourceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	req := request.ModifyGatewayRequest{
		UUID: d.Id(),
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
	}

	if d.HasChange("plan") {
		req.Plan = d.Get("plan").(string)
	}

	if d.HasChange("configured_status") {
		req.ConfiguredStatus = upcloud.GatewayConfiguredStatus(d.Get("configured_status").(string))
	}

	if d.HasChange("labels") {
		req.Labels = utils.LabelsMapToSlice(d.Get("labels").(map[string]interface{}))
	}

	svc := meta.(*service.Service)
	gw, err := svc.ModifyGateway(ctx, &req)
	if err != nil {
		return diag.FromErr(err)
	}

	return setGatewayResourceData(d, gw)
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*service.Service)
	if err := svc.DeleteGateway(ctx, &request.DeleteGatewayRequest{UUID: d.Id()}); err != nil {
		return diag.FromErr(err)
	}
	tflog.Info(ctx, "Gateway delete started", map[string]interface{}{"name": d.Get("name").(string), "uuid": d.Id()})

	// wait before continuing so that router can be deleted if needed
	diags := diag.FromErr(waitForGatewayToBeDeleted(ctx, svc, d.Id()))

	// Additionally wait some time so that all cleanup operations can finish
	time.Sleep(time.Second * cleanupWaitTimeSeconds)

	return diags
}

func setGatewayResourceData(d *schema.ResourceData, gw *upcloud.Gateway) (diags diag.Diagnostics) {
	if err := d.Set("name", gw.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("zone", gw.Zone); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("plan", gw.Plan); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("features", gw.Features); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("router", []map[string]interface{}{{"id": gw.Routers[0].UUID}}); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("labels", utils.LabelsSliceToMap(gw.Labels)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configured_status", gw.ConfiguredStatus); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("operational_state", gw.OperationalState); err != nil {
		return diag.FromErr(err)
	}

	var addresses []map[string]interface{}
	for _, address := range gw.Addresses {
		addresses = append(addresses, map[string]interface{}{
			"address": address.Address,
			"name":    address.Name,
		})
	}

	if err := d.Set("address", addresses); err != nil {
		return diag.FromErr(err)
	}

	var connections []string
	for _, conn := range gw.Connections {
		connections = append(connections, conn.Name)
	}

	if err := d.Set("connections", connections); err != nil {
		return diag.FromErr(err)
	}

	// This one is deprecated, can be removed later
	if err := d.Set("addresses", addresses); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func waitForGatewayToBeRunning(ctx context.Context, svc *service.Service, id string) (*upcloud.Gateway, error) {
	const maxRetries int = 500

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			gw, err := svc.GetGateway(ctx, &request.GetGatewayRequest{UUID: id})
			if err != nil {
				return nil, err
			}
			if gw.OperationalState == upcloud.GatewayOperationalStateRunning {
				return gw, nil
			}

			tflog.Info(ctx, "waiting for network gateway to be running", map[string]interface{}{"name": gw.Name, "state": gw.OperationalState})
		}
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("max retries (%d)reached while waiting for network gateway to be running", maxRetries)
}

func getGatewayDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	gw, err := svc.GetGateway(ctx, &request.GetGatewayRequest{UUID: id[0]})

	return map[string]interface{}{"resource": "gateway", "name": gw.Name, "state": gw.OperationalState}, err
}

func waitForGatewayToBeDeleted(ctx context.Context, svc *service.Service, id string) error {
	return utils.WaitForResourceToBeDeleted(ctx, svc, getGatewayDeleted, id)
}

var validateName = validation.ToDiagFunc(validation.All(
	validation.StringLenBetween(1, 64),
	validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9_-]+$"), "must contain only alphanumeric characters, hyphens, and underscores"),
))

var validateFeaturesElen = validation.ToDiagFunc(validation.StringInSlice([]string{
	string(upcloud.GatewayFeatureNAT),
	string(upcloud.GatewayFeatureVPN),
}, false))

var validateConfiguredStatus = validation.ToDiagFunc(validation.StringInSlice([]string{string(upcloud.GatewayConfiguredStatusStarted), string(upcloud.GatewayConfiguredStatusStopped)}, false))
