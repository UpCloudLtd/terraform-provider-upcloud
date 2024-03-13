package gateway

import (
	"context"
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
			"name": {
				Description:      "The name of the tunnel, should be unique within the connection",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateName,
			},
			"connection_id": {
				Description: "ID of the connection to which the tunnel belongs",
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

func resourceTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	svc := meta.(*service.Service)

	req, err := getCreateTunnelRequestFromSchema(d)
	if err != nil {
		return diag.FromErr(err)
	}

	tunnel, err := svc.CreateGatewayConnectionTunnel(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.MarshalID(req.ServiceUUID, req.ConnectionName, tunnel.Name))

	diags = append(diags, setTunnelResourceData(d, tunnel)...)
	if len(diags) > 0 {
		return diags
	}

	tflog.Info(ctx, "gateway tunnel created successfully", map[string]interface{}{"name": tunnel.Name, "service_uuid": req.ServiceUUID, "connection_name": req.ConnectionName})
	return diags
}

func resourceTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionName string
		tunnelName     string
	)

	err := utils.UnmarshalID(d.Id(), &serviceUUID, &connectionName, &tunnelName)
	if err != nil {
		return diag.FromErr(err)
	}

	tunnel, err := svc.GetGatewayConnectionTunnel(ctx, &request.GetGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionName: connectionName,
		Name:           tunnelName,
	})
	if err != nil {
		return utils.HandleResourceError(tunnelName, d, err)
	}

	d.SetId(utils.MarshalID(serviceUUID, connectionName, tunnel.Name))

	if err = d.Set("connection_id", utils.MarshalID(serviceUUID, connectionName)); err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, setTunnelResourceData(d, tunnel)...)
	if len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// var (
	// 	svc            = meta.(*service.Service)
	// 	serviceUUID    string
	// 	connectionName string
	// 	tunnelName     string
	// )

	// if err := utils.UnmarshalID(d.Id(), &serviceUUID, &connectionName, &tunnelName); err != nil {
	// 	return diag.FromErr(err)
	// }

	return diags
}

func resourceTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		svc            = meta.(*service.Service)
		serviceUUID    string
		connectionName string
		tunnelName     string
	)

	err := utils.UnmarshalID(d.Id(), &serviceUUID, &connectionName, &tunnelName)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "deleting gateway tunnel", map[string]interface{}{"name": tunnelName, "service_uuid": serviceUUID, "connection_name": connectionName})

	return diag.FromErr(svc.DeleteGatewayConnectionTunnel(ctx, &request.DeleteGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionName: connectionName,
		Name:           tunnelName,
	}))
}

func getCreateTunnelRequestFromSchema(d *schema.ResourceData) (*request.CreateGatewayConnectionTunnelRequest, error) {
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

	var serviceUUID, connectionName string
	err := utils.UnmarshalID(d.Get("connection_id").(string), &serviceUUID, &connectionName)
	if err != nil {
		return nil, err
	}

	result := &request.CreateGatewayConnectionTunnelRequest{
		ServiceUUID:    serviceUUID,
		ConnectionName: connectionName,
		Tunnel: request.GatewayTunnel{
			Name: d.Get("name").(string),
			LocalAddress: upcloud.GatewayTunnelLocalAddress{
				Name: d.Get("local_address_name").(string),
			},
			RemoteAddress: upcloud.GatewayTunnelRemoteAddress{
				Address: d.Get("remote_address").(string),
			},
			IPSec: upcloud.GatewayTunnelIPSec{
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
			},
		},
	}

	psk, authMethodIsPSK := d.GetOk("ipsec_auth_psk.0.psk")
	if authMethodIsPSK {
		result.Tunnel.IPSec.Authentication = upcloud.GatewayTunnelIPSecAuth{
			Authentication: upcloud.GatewayTunnelIPSecAuthTypePSK,
			PSK:            psk.(string),
		}
	}

	return result, nil
}

func setTunnelResourceData(d *schema.ResourceData, tunnel *upcloud.GatewayTunnel) diag.Diagnostics {
	var diags diag.Diagnostics

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

	if tunnel.IPSec.Authentication.Authentication == upcloud.GatewayTunnelIPSecAuthTypePSK {
		// We use slice of maps here because 'ipsec_auth_psk' is of type schema.TypeList, but with just one element
		ipsecAuthPSK := []map[string]interface{}{{
			"psk": "", // This value is only used during resource creation and should not really be stored in state
		}}

		if err := d.Set("ipsec_auth_psk", ipsecAuthPSK); err != nil {
			return diag.FromErr(err)
		}
	}

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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Id() != "" // create-only property
				},
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
