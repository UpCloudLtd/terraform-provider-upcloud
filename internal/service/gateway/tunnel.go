package gateway

import (
	"context"
	"fmt"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var allowedSecurityAlgorithms = []string{
	string(upcloud.GatewayIPSecAlgorithm_aes128),
	string(upcloud.GatewayIPSecAlgorithm_aes192),
	string(upcloud.GatewayIPSecAlgorithm_aes256),
	string(upcloud.GatewayIPSecAlgorithm_aes128gcm16),
	string(upcloud.GatewayIPSecAlgorithm_aes128gcm128),
	string(upcloud.GatewayIPSecAlgorithm_aes192gcm16),
	string(upcloud.GatewayIPSecAlgorithm_aes192gcm128),
	string(upcloud.GatewayIPSecAlgorithm_aes256gcm16),
	string(upcloud.GatewayIPSecAlgorithm_aes256gcm128),
}

var allowedIntegrityAlgorithms = []string{
	string(upcloud.GatewayIPSecIntegrityAlgorithm_aes128gmac),
	string(upcloud.GatewayIPSecIntegrityAlgorithm_aes256gmac),
	string(upcloud.GatewayIPSecIntegrityAlgorithm_sha1),
	string(upcloud.GatewayIPSecIntegrityAlgorithm_sha256),
	string(upcloud.GatewayIPSecIntegrityAlgorithm_sha384),
	string(upcloud.GatewayIPSecIntegrityAlgorithm_sha512),
}

var allowedDHGroups = []int{2, 5, 14, 15, 16, 18, 19, 20, 21, 24}

func ResourceTunnel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTunnelCreate,
		ReadContext:   resourceTunnelRead,
		UpdateContext: resourceTunnelUpdate,
		DeleteContext: resourceTunnelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"uuid": {
				Description: "The UUID of the tunnel",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description:      "The name of the tunnel, should be unique within the connection",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateName,
			},
			"connection_id": {
				Description: "ID of the upcloud_gateway_connection resource to which the tunnel belongs",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"local_address_name": {
				Description:      "Public (UpCloud) endpoint address of this tunnel",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateName,
			},
			"remote_address": {
				Description: "Remote public IP address of the tunnel",
				Type:        schema.TypeString,
				Required:    true,
			},
			"operational_state": {
				Description: "Tunnel's current operational, effective state",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipsec_auth_psk": {
				Description: "Configuration for authenticating with pre-shared key",
				ForceNew:    true,
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        ipsecAuthPSKSchema(),
			},
			"ipsec_properties": {
				Description: "IPsec configuration for the tunnel",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        ipsecPropertiesSchema(),
			},
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTunnelResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTunnelStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceTunnelResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "The name of the tunnel, should be unique within the connection",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateName,
			},
			"connection_id": {
				Description: "ID of the upcloud_gateway_connection resource to which the tunnel belongs",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"local_address_name": {
				Description:      "Public (UpCloud) endpoint address of this tunnel",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateName,
			},
			"remote_address": {
				Description: "Remote public IP address of the tunnel",
				Type:        schema.TypeString,
				Required:    true,
			},
			"operational_state": {
				Description: "Tunnel's current operational, effective state",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipsec_auth_psk": {
				Description: "Configuration for authenticating with pre-shared key",
				ForceNew:    true,
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        ipsecAuthPSKSchema(),
			},
			"ipsec_properties": {
				Description: "IPsec configuration for the tunnel",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        ipsecPropertiesSchema(),
			},
		},
	}
}

func resourceTunnelStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionName string
		connectionUUID string
		name           string
	)

	if err := utils.UnmarshalID(rawState["id"].(string), &serviceUUID, &connectionName, &name); err != nil {
		return rawState, err
	}

	conns, err := svc.GetGatewayConnections(ctx, &request.GetGatewayConnectionsRequest{ServiceUUID: serviceUUID})
	if err != nil {
		return rawState, err
	}

	for _, conn := range conns {
		if conn.Name == connectionName {
			connectionUUID = conn.UUID

			break
		}
	}

	tunnels, err := svc.GetGatewayConnectionTunnels(ctx, &request.GetGatewayConnectionTunnelsRequest{
		ServiceUUID:    serviceUUID,
		ConnectionUUID: connectionUUID,
	})
	if err != nil {
		return rawState, err
	}

	for _, tunnel := range tunnels {
		if tunnel.Name == rawState["name"].(string) {
			rawState["uuid"] = tunnel.UUID
			rawState["id"] = utils.MarshalID(serviceUUID, connectionUUID, tunnel.UUID)
			rawState["connection_id"] = utils.MarshalID(serviceUUID, connectionUUID)

			return rawState, nil
		}
	}

	return rawState, fmt.Errorf("tunnel by name %s not found", name)
}

func resourceTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionUUID string
	)

	if err := utils.UnmarshalID(d.Get("connection_id").(string), &serviceUUID, &connectionUUID); err != nil {
		return diag.FromErr(err)
	}

	ipsec, err := getIPSecRequestFieldsFromSchema(d)
	if err != nil {
		return diag.FromErr(err)
	}

	ipsecAuth, err := getIPSecAuthenticationFromSchema(d)
	if err != nil {
		return diag.FromErr(err)
	}

	ipsec.Authentication = ipsecAuth

	tunnel, err := svc.CreateGatewayConnectionTunnel(ctx, &request.CreateGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionUUID: connectionUUID,
		Tunnel: request.GatewayTunnel{
			Name: d.Get("name").(string),
			LocalAddress: upcloud.GatewayTunnelLocalAddress{
				Name: d.Get("local_address_name").(string),
			},
			RemoteAddress: upcloud.GatewayTunnelRemoteAddress{
				Address: d.Get("remote_address").(string),
			},
			IPSec: ipsec,
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceUUID, connectionUUID, tunnel.UUID))

	diags = append(diags, setTunnelResourceData(d, tunnel)...)
	if len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "gateway tunnel created successfully", map[string]interface{}{"uuid": tunnel.UUID, "service_uuid": serviceUUID, "connection_uuid": connectionUUID})
	return diags
}

func resourceTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionUUID string
		uuid           string
	)

	err := utils.UnmarshalID(d.Id(), &serviceUUID, &connectionUUID, &uuid)
	if err != nil {
		return diag.FromErr(err)
	}

	tunnel, err := svc.GetGatewayConnectionTunnel(ctx, &request.GetGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionUUID: connectionUUID,
		UUID:           uuid,
	})
	if err != nil {
		return utils.HandleResourceError(uuid, d, err)
	}

	d.SetId(utils.MarshalID(serviceUUID, connectionUUID, tunnel.UUID))

	if err = d.Set("connection_id", utils.MarshalID(serviceUUID, connectionUUID)); err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, setTunnelResourceData(d, tunnel)...)
	if len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionUUID string
		uuid           string
	)

	if err := utils.UnmarshalID(d.Id(), &serviceUUID, &connectionUUID, &uuid); err != nil {
		return diag.FromErr(err)
	}

	req := request.ModifyGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionUUID: connectionUUID,
		UUID:           uuid,
		Tunnel: request.ModifyGatewayTunnel{
			// We don't allow updating the tunnel name in TF, but as of now it is a required parameter in the request payload (due to some bug)
			// TODO: remove once API allows modification requests without the name
			Name: d.Get("name").(string),
		},
	}

	if d.HasChange("ipsec_properties") {
		ipsec, err := getIPSecRequestFieldsFromSchema(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req.Tunnel.IPSec = &ipsec
	}

	if d.HasChange("local_address_name") {
		req.Tunnel.LocalAddress = &upcloud.GatewayTunnelLocalAddress{
			Name: d.Get("local_address_name").(string),
		}
	}

	if d.HasChange("remote_address") {
		req.Tunnel.RemoteAddress = &upcloud.GatewayTunnelRemoteAddress{
			Address: d.Get("remote_address").(string),
		}
	}

	tunnel, err := svc.ModifyGatewayConnectionTunnel(ctx, &req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(serviceUUID, connectionUUID, tunnel.UUID))

	if diags = append(diags, setTunnelResourceData(d, tunnel)...); len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "gateway tunnel updated", map[string]interface{}{"uuid": tunnel.UUID, "service_uuid": serviceUUID, "connection_uuid": connectionUUID})
	return diags
}

func resourceTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionUUID string
		uuid           string
	)

	err := utils.UnmarshalID(d.Id(), &serviceUUID, &connectionUUID, &uuid)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "deleting gateway tunnel", map[string]interface{}{"uuid": uuid, "service_uuid": serviceUUID, "connection_uuid": connectionUUID})

	return diag.FromErr(svc.DeleteGatewayConnectionTunnel(ctx, &request.DeleteGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionUUID: connectionUUID,
		UUID:           uuid,
	}))
}

func getIPSecRequestFieldsFromSchema(d *schema.ResourceData) (upcloud.GatewayTunnelIPSec, error) {
	phase1Algs := []upcloud.GatewayIPSecAlgorithm{}
	for _, alg := range d.Get("ipsec_properties.0.phase1_algorithms").(*schema.Set).List() {
		phase1Algs = append(phase1Algs, upcloud.GatewayIPSecAlgorithm(alg.(string)))
	}

	phase1DHGroupNumbers := []int{}
	for _, num := range d.Get("ipsec_properties.0.phase1_dh_group_numbers").(*schema.Set).List() {
		phase1DHGroupNumbers = append(phase1DHGroupNumbers, num.(int))
	}

	phase1IntegrityAlgs := []upcloud.GatewayIPSecIntegrityAlgorithm{}
	for _, alg := range d.Get("ipsec_properties.0.phase1_integrity_algorithms").(*schema.Set).List() {
		phase1IntegrityAlgs = append(phase1IntegrityAlgs, upcloud.GatewayIPSecIntegrityAlgorithm(alg.(string)))
	}

	phase2Algs := []upcloud.GatewayIPSecAlgorithm{}
	for _, alg := range d.Get("ipsec_properties.0.phase2_algorithms").(*schema.Set).List() {
		phase2Algs = append(phase2Algs, upcloud.GatewayIPSecAlgorithm(alg.(string)))
	}

	phase2DHGroupNumbers := []int{}
	for _, num := range d.Get("ipsec_properties.0.phase2_dh_group_numbers").(*schema.Set).List() {
		phase2DHGroupNumbers = append(phase2DHGroupNumbers, num.(int))
	}

	phase2IntegrityAlgs := []upcloud.GatewayIPSecIntegrityAlgorithm{}
	for _, alg := range d.Get("ipsec_properties.0.phase2_integrity_algorithms").(*schema.Set).List() {
		phase2IntegrityAlgs = append(phase2IntegrityAlgs, upcloud.GatewayIPSecIntegrityAlgorithm(alg.(string)))
	}

	return upcloud.GatewayTunnelIPSec{
		ChildRekeyTime:            d.Get("ipsec_properties.0.child_rekey_time").(int),
		DPDDelay:                  d.Get("ipsec_properties.0.dpd_delay").(int),
		DPDTimeout:                d.Get("ipsec_properties.0.dpd_timeout").(int),
		IKELifetime:               d.Get("ipsec_properties.0.ike_lifetime").(int),
		RekeyTime:                 d.Get("ipsec_properties.0.rekey_time").(int),
		Phase1Algorithms:          phase1Algs,
		Phase1DHGroupNumbers:      phase1DHGroupNumbers,
		Phase1IntegrityAlgorithms: phase1IntegrityAlgs,
		Phase2Algorithms:          phase2Algs,
		Phase2DHGroupNumbers:      phase2DHGroupNumbers,
		Phase2IntegrityAlgorithms: phase2IntegrityAlgs,
	}, nil
}

func getIPSecAuthenticationFromSchema(d *schema.ResourceData) (upcloud.GatewayTunnelIPSecAuth, error) {
	result := upcloud.GatewayTunnelIPSecAuth{}

	if psk, authMethodIsPSK := d.GetOk("ipsec_auth_psk.0.psk"); authMethodIsPSK {
		result.Authentication = upcloud.GatewayTunnelIPSecAuthTypePSK
		result.PSK = psk.(string)
		return result, nil
	}

	// Put more authentication methods here once supported

	return result, fmt.Errorf("tunnel IPsec authentication method not recognized")
}

func setTunnelResourceData(d *schema.ResourceData, tunnel *upcloud.GatewayTunnel) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := d.Set("uuid", tunnel.UUID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", tunnel.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("local_address_name", tunnel.LocalAddress.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("remote_address", tunnel.RemoteAddress.Address); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("operational_state", string(tunnel.OperationalState)); err != nil {
		return diag.FromErr(err)
	}

	// We do not set ipsec_auth_psk block as API don't return PSK in API responses
	// We rely on TF core to just set it to state and track changes

	// Again, slice of maps because it's a schema.TypeList
	ipsecProperties := []map[string]interface{}{{
		"child_rekey_time":            tunnel.IPSec.ChildRekeyTime,
		"dpd_delay":                   tunnel.IPSec.DPDDelay,
		"dpd_timeout":                 tunnel.IPSec.DPDTimeout,
		"ike_lifetime":                tunnel.IPSec.IKELifetime,
		"rekey_time":                  tunnel.IPSec.RekeyTime,
		"phase1_algorithms":           tunnel.IPSec.Phase1Algorithms,
		"phase1_dh_group_numbers":     tunnel.IPSec.Phase1DHGroupNumbers,
		"phase1_integrity_algorithms": tunnel.IPSec.Phase1IntegrityAlgorithms,
		"phase2_algorithms":           tunnel.IPSec.Phase2Algorithms,
		"phase2_dh_group_numbers":     tunnel.IPSec.Phase2DHGroupNumbers,
		"phase2_integrity_algorithms": tunnel.IPSec.Phase2IntegrityAlgorithms,
	}}

	if err := d.Set("ipsec_properties", ipsecProperties); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func ipsecAuthPSKSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"psk": {
				Description: "The pre-shared key. This value is only used during resource creation and is not returned in the state. It is not possible to update this value. If you need to update it, delete the connection and create a new one.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(
					validation.StringLenBetween(8, 64),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z1-9_.][a-zA-Z0-9_.]+$"), "must contain only alphanumeric characters, underscores, and dots"),
				)),
			},
		},
	}
}

func ipsecPropertiesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"child_rekey_time": {
				Description: "IKE child SA rekey time in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"dpd_delay": {
				Description: "Delay before sending Dead Peer Detection packets if no traffic is detected, in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"dpd_timeout": {
				Description: "Timeout period for DPD reply before considering the peer to be dead, in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"ike_lifetime": {
				Description: "Maximum IKE SA lifetime in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"phase1_algorithms": {
				Description: "List of Phase 1: Proposal algorithms.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(allowedSecurityAlgorithms, false)),
				},
			},
			"phase1_dh_group_numbers": {
				Description: "List of Phase 1 Diffie-Hellman group numbers.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeInt,
					ValidateDiagFunc: validation.ToDiagFunc(validation.IntInSlice(allowedDHGroups)),
				},
			},
			"phase1_integrity_algorithms": {
				Description: "List of Phase 1 integrity algorithms.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(allowedIntegrityAlgorithms, false)),
				},
			},
			"phase2_algorithms": {
				Description: "List of Phase 2: Security Association algorithms.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(allowedSecurityAlgorithms, false)),
				},
			},
			"phase2_dh_group_numbers": {
				Description: "List of Phase 2 Diffie-Hellman group numbers.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeInt,
					ValidateDiagFunc: validation.ToDiagFunc(validation.IntInSlice(allowedDHGroups)),
				},
			},
			"phase2_integrity_algorithms": {
				Description: "List of Phase 2 integrity algorithms.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(allowedIntegrityAlgorithms, false)),
				},
			},
			"rekey_time": {
				Description: "IKE SA rekey time in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}
