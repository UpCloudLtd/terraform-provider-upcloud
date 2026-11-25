package kubernetes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	v9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &kubernetesClusterAuthenticationResource{}
	_ resource.ResourceWithConfigure   = &kubernetesClusterAuthenticationResource{}
	_ resource.ResourceWithImportState = &kubernetesClusterAuthenticationResource{}
)

func NewKubernetesClusterAuthenticationResource() resource.Resource {
	return &kubernetesClusterAuthenticationResource{}
}

type kubernetesClusterAuthenticationResource struct {
	client *v9.ClientWithResponses
}

func (r *kubernetesClusterAuthenticationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster_authentication"
}

func (r *kubernetesClusterAuthenticationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetV9ClientFromProviderData(req.ProviderData)
}

type clusterAuthenticationModel struct {
	ID                   types.String               `tfsdk:"id"`
	Cluster              types.String               `tfsdk:"cluster"`
	Name                 types.String               `tfsdk:"name"`
	IssuerURL            types.String               `tfsdk:"issuer_url"`
	DiscoveryURL         types.String               `tfsdk:"discovery_url"`
	Audiences            []types.String             `tfsdk:"audiences"`
	AudienceMatchPolicy  types.String               `tfsdk:"audience_match_policy"`
	CertificateAuthority types.String               `tfsdk:"certificate_authority"`
	ClaimMappings        *claimMappingsModel        `tfsdk:"claim_mappings"`
	ClaimValidationRules []claimValidationRuleModel `tfsdk:"claim_validation_rules"`
	UserValidationRules  []userValidationRuleModel  `tfsdk:"user_validation_rules"`
}

type claimMappingsModel struct {
	Username *prefixedClaimOrExpressionModel `tfsdk:"username"`
	Groups   *prefixedClaimOrExpressionModel `tfsdk:"groups"`
	UID      *claimOrExpressionModel         `tfsdk:"uid"`
	Extra    []extraMappingModel             `tfsdk:"extra"`
}

type prefixedClaimOrExpressionModel struct {
	Claim      types.String `tfsdk:"claim"`
	Prefix     types.String `tfsdk:"prefix"`
	Expression types.String `tfsdk:"expression"`
}

type claimOrExpressionModel struct {
	Claim      types.String `tfsdk:"claim"`
	Expression types.String `tfsdk:"expression"`
}

type extraMappingModel struct {
	Key             types.String `tfsdk:"key"`
	ValueExpression types.String `tfsdk:"value_expression"`
}

type claimValidationRuleModel struct {
	Claim         types.String `tfsdk:"claim"`
	RequiredValue types.String `tfsdk:"required_value"`
	Expression    types.String `tfsdk:"expression"`
	Message       types.String `tfsdk:"message"`
}

type userValidationRuleModel struct {
	Expression types.String `tfsdk:"expression"`
	Message    types.String `tfsdk:"message"`
}

func siblingPath(name string) path.Expression {
	return path.MatchRelative().AtParent().AtName(name)
}

func claimOrExpressionValidators() []validator.String {
	return []validator.String{stringvalidator.ExactlyOneOf(siblingPath("expression"))}
}

func (r *kubernetesClusterAuthenticationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents an OIDC authentication issuer configuration of a [Managed Kubernetes](https://upcloud.com/products/managed-kubernetes) cluster. The cluster API server accepts tokens from the configured issuers in addition to the default authentication methods.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed ID of the authentication issuer configuration. This is a combination of the cluster UUID and the configuration name, separated with a `/`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster": schema.StringAttribute{
				MarkdownDescription: idDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the authentication issuer configuration. Needs to be unique within a cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, resourceNameMaxLength),
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name should only contain lowercase alphanumeric characters and dashes (a-z, 0-9, -). Name should not start or end with a dash. Regular expresion used to check validation: %s", resourceNameRegexp)),
				},
			},
			"issuer_url": schema.StringAttribute{
				MarkdownDescription: "URL of the OIDC issuer, e.g. `https://accounts.myoidcprovider.com`.",
				Required:            true,
			},
			"discovery_url": schema.StringAttribute{
				MarkdownDescription: "URL of the OIDC discovery document, e.g. `https://accounts.myoidcprovider.com/.well-known/openid-configuration`. Use only if the discovery document is not served from the default location under `issuer_url`.",
				Optional:            true,
			},
			"audiences": schema.ListAttribute{
				MarkdownDescription: "Accepted audiences for tokens from this issuer. By default, the first audience is used as the OIDC client ID in the generated kubeconfigs.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"audience_match_policy": schema.StringAttribute{
				MarkdownDescription: "Audience match policy. Set to `MatchAny` to accept a token that has any of the configured `audiences`. Required if `audiences` contains multiple values. Leave out to use the default policy.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(v9.KubernetesAuthenticationIssuerConfigAudienceMatchPolicyMatchAny)),
				},
			},
			"certificate_authority": schema.StringAttribute{
				MarkdownDescription: "Base64-encoded PEM certificate authority bundle used to validate the TLS certificate of the issuer.",
				Optional:            true,
			},
			"claim_mappings": schema.SingleNestedAttribute{
				MarkdownDescription: "Rules for mapping token claims to Kubernetes user attributes.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.SingleNestedAttribute{
						MarkdownDescription: "Mapping for the Kubernetes username. Use either `claim` with `prefix` or `expression`.",
						Optional:            true,
						Attributes:          prefixedClaimOrExpressionAttributes(),
						// Validators of nested attributes run only when `claim_mappings` is set, so this checks that the object is not empty.
						Validators: []validator.Object{
							objectvalidator.AtLeastOneOf(
								siblingPath("groups"),
								siblingPath("uid"),
								siblingPath("extra"),
							),
						},
					},
					"groups": schema.SingleNestedAttribute{
						MarkdownDescription: "Mapping for the Kubernetes groups. Use either `claim` with `prefix` or `expression`.",
						Optional:            true,
						Attributes:          prefixedClaimOrExpressionAttributes(),
					},
					"uid": schema.SingleNestedAttribute{
						MarkdownDescription: "Mapping for the Kubernetes user UID. Use either `claim` or `expression`.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"claim": schema.StringAttribute{
								MarkdownDescription: "JWT claim to use. Mutually exclusive with `expression`.",
								Optional:            true,
								Validators:          claimOrExpressionValidators(),
							},
							"expression": schema.StringAttribute{
								MarkdownDescription: "CEL expression producing the attribute value, e.g. `claims.sub`. Mutually exclusive with `claim`.",
								Optional:            true,
							},
						},
					},
					"extra": schema.ListNestedAttribute{
						MarkdownDescription: "Mappings for extra user attributes.",
						Optional:            true,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									MarkdownDescription: "Domain-prefixed, lowercase key for the extra attribute, e.g. `example.org/foo`.",
									Required:            true,
								},
								"value_expression": schema.StringAttribute{
									MarkdownDescription: "CEL expression producing the extra attribute value (string or string array), e.g. `claims.some_claim`.",
									Required:            true,
								},
							},
						},
					},
				},
			},
			"claim_validation_rules": schema.ListNestedAttribute{
				MarkdownDescription: "Rules that the token claims need to fulfill for the token to be accepted. Use either `claim` with `required_value` or `expression` with `message` in each rule.",
				Optional:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"claim": schema.StringAttribute{
							MarkdownDescription: "Name of the claim to validate. Mutually exclusive with `expression`.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(siblingPath("expression")),
							},
						},
						"required_value": schema.StringAttribute{
							MarkdownDescription: "Value that the claim must have, e.g. `upcloud.com`. Mutually exclusive with `expression`.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(siblingPath("expression")),
							},
						},
						"expression": schema.StringAttribute{
							MarkdownDescription: "CEL expression that must evaluate to `true`, e.g. `claims.hd == 'upcloud.com'`. Mutually exclusive with `claim` and `required_value`.",
							Optional:            true,
						},
						"message": schema.StringAttribute{
							MarkdownDescription: "Message to show when the validation fails.",
							Optional:            true,
						},
					},
				},
			},
			"user_validation_rules": schema.ListNestedAttribute{
				MarkdownDescription: "CEL rules that are applied to the final authenticated user.",
				Optional:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"expression": schema.StringAttribute{
							MarkdownDescription: "CEL expression that must evaluate to `true`, e.g. `!user.username.startsWith('system:')`.",
							Required:            true,
						},
						"message": schema.StringAttribute{
							MarkdownDescription: "Message to show when the validation fails.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func prefixedClaimOrExpressionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"claim": schema.StringAttribute{
			MarkdownDescription: "JWT claim to use, e.g. `email`. Requires `prefix`. Mutually exclusive with `expression`.",
			Optional:            true,
			Validators: append(claimOrExpressionValidators(),
				stringvalidator.AlsoRequires(siblingPath("prefix")),
			),
		},
		"prefix": schema.StringAttribute{
			MarkdownDescription: "Prefix prepended to the claim's value, e.g. `oidc:`. Set to an empty string to not use a prefix. Mutually exclusive with `expression`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(siblingPath("expression")),
			},
		},
		"expression": schema.StringAttribute{
			MarkdownDescription: "CEL expression producing the attribute value, e.g. `claims.email`. Mutually exclusive with `claim` and `prefix`.",
			Optional:            true,
		},
	}
}

func clusterAuthenticationError(summary string, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(summary, utils.ErrorDiagnosticDetail(err))
	return diags
}

func clusterAuthenticationStatusError(summary, status string, body []byte) diag.Diagnostics {
	var diags diag.Diagnostics
	detail := fmt.Sprintf("API returned unexpected status %s", status)
	if len(body) > 0 {
		detail = fmt.Sprintf("%s. Response: %s", detail, string(body))
	}
	diags.AddError(summary, detail)
	return diags
}

func parseClusterUUID(value string) (uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	id, err := uuid.Parse(value)
	if err != nil {
		diags.AddAttributeError(path.Root("cluster"), "Invalid cluster UUID", utils.ErrorDiagnosticDetail(err))
	}
	return id, diags
}

func (r *kubernetesClusterAuthenticationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterAuthenticationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUUID, diags := parseClusterUUID(plan.Cluster.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.CreateClusterAuthenticationWithResponse(ctx, clusterUUID, expandClusterAuthentication(plan))
	if err != nil {
		resp.Diagnostics.Append(clusterAuthenticationError("Unable to create Kubernetes cluster authentication", err)...)
		return
	}
	if apiResp.StatusCode() != http.StatusCreated {
		resp.Diagnostics.Append(clusterAuthenticationStatusError("Unable to create Kubernetes cluster authentication", apiResp.Status(), apiResp.Body)...)
		return
	}

	plan.ID = types.StringValue(utils.MarshalID(plan.Cluster.ValueString(), plan.Name.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kubernetesClusterAuthenticationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterAuthenticationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cluster, name string
	resp.Diagnostics.Append(utils.UnmarshalIDDiag(state.ID.ValueString(), &cluster, &name)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUUID, diags := parseClusterUUID(cluster)
	if diags.HasError() {
		resp.Diagnostics.AddError("Invalid Kubernetes cluster authentication ID", fmt.Sprintf("Expected ID in the format `<cluster UUID>/<name>`, got: %s", state.ID.ValueString()))
		return
	}

	apiResp, err := r.client.GetClusterAuthenticationWithResponse(ctx, clusterUUID, name)
	if err != nil {
		resp.Diagnostics.Append(clusterAuthenticationError("Unable to read Kubernetes cluster authentication", err)...)
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.StatusCode() != http.StatusOK || apiResp.JSON200 == nil {
		resp.Diagnostics.Append(clusterAuthenticationStatusError("Unable to read Kubernetes cluster authentication", apiResp.Status(), apiResp.Body)...)
		return
	}

	state.Cluster = types.StringValue(cluster)
	flattenClusterAuthentication(&state, apiResp.JSON200)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *kubernetesClusterAuthenticationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterAuthenticationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUUID, diags := parseClusterUUID(plan.Cluster.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.UpdateClusterAuthenticationWithResponse(ctx, clusterUUID, plan.Name.ValueString(), expandClusterAuthentication(plan))
	if err != nil {
		resp.Diagnostics.Append(clusterAuthenticationError("Unable to update Kubernetes cluster authentication", err)...)
		return
	}
	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.Append(clusterAuthenticationStatusError("Unable to update Kubernetes cluster authentication", apiResp.Status(), apiResp.Body)...)
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kubernetesClusterAuthenticationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterAuthenticationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUUID, diags := parseClusterUUID(state.Cluster.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.DeleteClusterAuthenticationWithResponse(ctx, clusterUUID, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(clusterAuthenticationError("Unable to delete Kubernetes cluster authentication", err)...)
		return
	}
	if code := apiResp.StatusCode(); code >= http.StatusMultipleChoices && code != http.StatusNotFound {
		resp.Diagnostics.Append(clusterAuthenticationStatusError("Unable to delete Kubernetes cluster authentication", apiResp.Status(), apiResp.Body)...)
	}
}

func (r *kubernetesClusterAuthenticationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func expandClusterAuthentication(m clusterAuthenticationModel) v9.KubernetesAuthenticationIssuerConfig {
	audiences := make([]string, 0, len(m.Audiences))
	for _, a := range m.Audiences {
		audiences = append(audiences, a.ValueString())
	}

	config := v9.KubernetesAuthenticationIssuerConfig{
		Name:                 m.Name.ValueString(),
		IssuerUrl:            m.IssuerURL.ValueString(),
		DiscoveryUrl:         m.DiscoveryURL.ValueStringPointer(),
		Audiences:            audiences,
		CertificateAuthority: m.CertificateAuthority.ValueStringPointer(),
	}

	if !m.AudienceMatchPolicy.IsNull() {
		policy := v9.KubernetesAuthenticationIssuerConfigAudienceMatchPolicy(m.AudienceMatchPolicy.ValueString())
		config.AudienceMatchPolicy = &policy
	}

	if m.ClaimMappings != nil {
		config.ClaimMappings = expandClaimMappings(m.ClaimMappings)
	}

	if len(m.ClaimValidationRules) > 0 {
		rules := make([]v9.KubernetesClaimValidationRule, 0, len(m.ClaimValidationRules))
		for _, rule := range m.ClaimValidationRules {
			rules = append(rules, v9.KubernetesClaimValidationRule{
				Claim:         rule.Claim.ValueStringPointer(),
				RequiredValue: rule.RequiredValue.ValueStringPointer(),
				Expression:    rule.Expression.ValueStringPointer(),
				Message:       rule.Message.ValueStringPointer(),
			})
		}
		config.ClaimValidationRules = &rules
	}

	if len(m.UserValidationRules) > 0 {
		rules := make([]v9.KubernetesUserValidationRule, 0, len(m.UserValidationRules))
		for _, rule := range m.UserValidationRules {
			rules = append(rules, v9.KubernetesUserValidationRule{
				Expression: rule.Expression.ValueString(),
				Message:    rule.Message.ValueStringPointer(),
			})
		}
		config.UserValidationRules = &rules
	}

	return config
}

func expandClaimMappings(m *claimMappingsModel) *v9.KubernetesClaimMappings {
	mappings := &v9.KubernetesClaimMappings{}

	if m.Username != nil {
		mappings.Username = expandPrefixedClaimOrExpression(m.Username)
	}
	if m.Groups != nil {
		mappings.Groups = expandPrefixedClaimOrExpression(m.Groups)
	}
	if m.UID != nil {
		mappings.Uid = &v9.KubernetesClaimOrExpression{
			Claim:      m.UID.Claim.ValueStringPointer(),
			Expression: m.UID.Expression.ValueStringPointer(),
		}
	}
	if len(m.Extra) > 0 {
		extra := make([]v9.KubernetesExtraMapping, 0, len(m.Extra))
		for _, e := range m.Extra {
			extra = append(extra, v9.KubernetesExtraMapping{
				Key:             e.Key.ValueString(),
				ValueExpression: e.ValueExpression.ValueString(),
			})
		}
		mappings.Extra = &extra
	}

	return mappings
}

func expandPrefixedClaimOrExpression(m *prefixedClaimOrExpressionModel) *v9.KubernetesPrefixedClaimOrExpression {
	return &v9.KubernetesPrefixedClaimOrExpression{
		Claim:      m.Claim.ValueStringPointer(),
		Prefix:     m.Prefix.ValueStringPointer(),
		Expression: m.Expression.ValueStringPointer(),
	}
}

// unsetString maps an empty or missing API string to null.
func unsetString(v *string) types.String {
	if v == nil || *v == "" {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func flattenClusterAuthentication(m *clusterAuthenticationModel, api *v9.KubernetesAuthenticationIssuerConfig) {
	m.ID = types.StringValue(utils.MarshalID(m.Cluster.ValueString(), api.Name))
	m.Name = types.StringValue(api.Name)
	m.IssuerURL = types.StringValue(api.IssuerUrl)
	m.DiscoveryURL = unsetString(api.DiscoveryUrl)
	m.CertificateAuthority = unsetString(api.CertificateAuthority)

	m.Audiences = make([]types.String, 0, len(api.Audiences))
	for _, a := range api.Audiences {
		m.Audiences = append(m.Audiences, types.StringValue(a))
	}

	m.AudienceMatchPolicy = types.StringNull()
	if api.AudienceMatchPolicy != nil && *api.AudienceMatchPolicy != v9.KubernetesAuthenticationIssuerConfigAudienceMatchPolicyEmpty {
		m.AudienceMatchPolicy = types.StringValue(string(*api.AudienceMatchPolicy))
	}

	m.ClaimMappings = flattenClaimMappings(api.ClaimMappings)

	m.ClaimValidationRules = nil
	if api.ClaimValidationRules != nil {
		for _, rule := range *api.ClaimValidationRules {
			m.ClaimValidationRules = append(m.ClaimValidationRules, claimValidationRuleModel{
				Claim:         unsetString(rule.Claim),
				RequiredValue: unsetString(rule.RequiredValue),
				Expression:    unsetString(rule.Expression),
				Message:       unsetString(rule.Message),
			})
		}
	}

	m.UserValidationRules = nil
	if api.UserValidationRules != nil {
		for _, rule := range *api.UserValidationRules {
			m.UserValidationRules = append(m.UserValidationRules, userValidationRuleModel{
				Expression: types.StringValue(rule.Expression),
				Message:    unsetString(rule.Message),
			})
		}
	}
}

func flattenClaimMappings(api *v9.KubernetesClaimMappings) *claimMappingsModel {
	if api == nil {
		return nil
	}

	m := &claimMappingsModel{
		Username: flattenPrefixedClaimOrExpression(api.Username),
		Groups:   flattenPrefixedClaimOrExpression(api.Groups),
	}

	if api.Uid != nil && (api.Uid.Claim != nil || api.Uid.Expression != nil) {
		m.UID = &claimOrExpressionModel{
			Claim:      unsetString(api.Uid.Claim),
			Expression: unsetString(api.Uid.Expression),
		}
	}

	if api.Extra != nil {
		for _, e := range *api.Extra {
			m.Extra = append(m.Extra, extraMappingModel{
				Key:             types.StringValue(e.Key),
				ValueExpression: types.StringValue(e.ValueExpression),
			})
		}
	}

	if m.Username == nil && m.Groups == nil && m.UID == nil && m.Extra == nil {
		return nil
	}
	return m
}

func flattenPrefixedClaimOrExpression(api *v9.KubernetesPrefixedClaimOrExpression) *prefixedClaimOrExpressionModel {
	if api == nil || (unsetString(api.Claim).IsNull() && unsetString(api.Expression).IsNull()) {
		return nil
	}

	m := &prefixedClaimOrExpressionModel{
		Claim:      unsetString(api.Claim),
		Expression: unsetString(api.Expression),
		Prefix:     types.StringPointerValue(api.Prefix),
	}
	// Prefix is required, but can be empty, when claim is used.
	if !m.Claim.IsNull() && api.Prefix == nil {
		m.Prefix = types.StringValue("")
	}
	return m
}
