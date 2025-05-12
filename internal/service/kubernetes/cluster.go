package kubernetes

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &kubernetesClusterResource{}
	_ resource.ResourceWithConfigure   = &kubernetesClusterResource{}
	_ resource.ResourceWithImportState = &kubernetesClusterResource{}
)

func NewKubernetesClusterResource() resource.Resource {
	return &kubernetesClusterResource{}
}

type kubernetesClusterResource struct {
	client *service.Service
}

func (r *kubernetesClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

// Configure adds the provider configured client to the resource.
func (r *kubernetesClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type kubernetesClusterModel struct {
	ControlPlaneIPFilter types.Set    `tfsdk:"control_plane_ip_filter"`
	ID                   types.String `tfsdk:"id"`
	Labels               types.Map    `tfsdk:"labels"`
	Name                 types.String `tfsdk:"name"`
	Network              types.String `tfsdk:"network"`
	NetworkCIDR          types.String `tfsdk:"network_cidr"`
	NodeGroups           types.List   `tfsdk:"node_groups"`
	Plan                 types.String `tfsdk:"plan"`
	PrivateNodeGroups    types.Bool   `tfsdk:"private_node_groups"`
	State                types.String `tfsdk:"state"`
	StorageEncryption    types.String `tfsdk:"storage_encryption"`
	Version              types.String `tfsdk:"version"`
	UpgradeStrategyType  types.String `tfsdk:"upgrade_strategy_type"`
	Zone                 types.String `tfsdk:"zone"`
}

func (r *kubernetesClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents a [Managed Kubernetes](https://upcloud.com/products/managed-kubernetes) cluster.",
		Attributes: map[string]schema.Attribute{
			"control_plane_ip_filter": schema.SetAttribute{
				MarkdownDescription: controlPlaneIPFilterDescription,
				Required:            true,
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: idDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": utils.LabelsAttribute("cluster"),
			"name": schema.StringAttribute{
				MarkdownDescription: nameDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(resourceNameMaxLength),
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name should only contain lowercase alphanumeric characters and dashes (a-z, 0-9, -). Name should not start or end with a dash. Regular expresion used to check validation: %s", resourceNameRegexp)),
				},
			},
			"network": schema.StringAttribute{
				MarkdownDescription: networkDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_cidr": schema.StringAttribute{
				MarkdownDescription: networkCIDRDescription,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"node_groups": schema.ListAttribute{
				MarkdownDescription: nodeGroupNamesDescription,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: planDescription,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_node_groups": schema.BoolAttribute{
				MarkdownDescription: privateNodeGroupsDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: stateDescription,
				Computed:            true,
			},
			"storage_encryption": schema.StringAttribute{
				Description: clusterStorageEncryptionDescription,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(string(upcloud.StorageEncryptionDataAtRest)),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: versionDescription,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"upgrade_strategy_type": schema.StringAttribute{
				MarkdownDescription: upgradeStrategyDescription,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.KubernetesUpgradeStrategyManual),
						string(upcloud.KubernetesUpgradeStrategyRollingUpdate),
					),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: zoneDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func setClusterValues(ctx context.Context, data *kubernetesClusterModel, cluster *upcloud.KubernetesCluster) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	nodeGroupNames := make([]string, len(cluster.NodeGroups))
	for i, nodeGroup := range cluster.NodeGroups {
		nodeGroupNames[i] = nodeGroup.Name
	}

	data.ControlPlaneIPFilter, diags = types.SetValueFrom(ctx, types.StringType, cluster.ControlPlaneIPFilter)
	respDiagnostics.Append(diags...)

	data.ID = types.StringValue(cluster.UUID)

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(cluster.Labels))
	respDiagnostics.Append(diags...)

	data.Name = types.StringValue(cluster.Name)
	data.Network = types.StringValue(cluster.Network)
	data.NetworkCIDR = types.StringValue(cluster.NetworkCIDR)
	data.NodeGroups, respDiagnostics = types.ListValueFrom(ctx, types.StringType, nodeGroupNames)
	data.Plan = types.StringValue(cluster.Plan)
	data.PrivateNodeGroups = types.BoolValue(cluster.PrivateNodeGroups)
	data.State = types.StringValue(string(cluster.State))

	if cluster.StorageEncryption == "" {
		data.StorageEncryption = types.StringNull()
	} else {
		data.StorageEncryption = types.StringValue(string(cluster.StorageEncryption))
	}

	data.Version = types.StringValue(cluster.Version)
	data.Zone = types.StringValue(cluster.Zone)

	return respDiagnostics
}

func (r *kubernetesClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ipFilter []string
	resp.Diagnostics.Append(data.ControlPlaneIPFilter.ElementsAs(ctx, &ipFilter, false)...)

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	apiReq := request.CreateKubernetesClusterRequest{
		ControlPlaneIPFilter: ipFilter,
		Name:                 data.Name.ValueString(),
		Network:              data.Network.ValueString(),
		Labels:               utils.LabelsMapToSlice(labels),
		Plan:                 data.Plan.ValueString(),
		PrivateNodeGroups:    data.PrivateNodeGroups.ValueBool(),
		StorageEncryption:    upcloud.StorageEncryption(data.StorageEncryption.ValueString()),
		Version:              data.Version.ValueString(),
		Zone:                 data.Zone.ValueString(),
	}

	cluster, err := r.client.CreateKubernetesCluster(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Kubernetes cluster",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(cluster.UUID)

	cluster, err = r.client.WaitForKubernetesClusterState(ctx, &request.WaitForKubernetesClusterStateRequest{
		DesiredState: upcloud.KubernetesClusterStateRunning,
		UUID:         cluster.UUID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for Kubernetes cluster to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setClusterValues(ctx, &data, cluster)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	cluster, err := r.client.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{
		UUID: data.ID.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read Kubernetes cluster details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setClusterValues(ctx, &data, cluster)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan kubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var labels map[string]string
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	}
	labelsSlice := utils.NilAsEmptyList(utils.LabelsMapToSlice(labels))

	var ipFilter []string
	resp.Diagnostics.Append(plan.ControlPlaneIPFilter.ElementsAs(ctx, &ipFilter, false)...)

	apiReq := &request.ModifyKubernetesClusterRequest{
		ClusterUUID: plan.ID.ValueString(),
		Cluster: request.ModifyKubernetesCluster{
			ControlPlaneIPFilter: &ipFilter,
			Labels:               &labelsSlice,
		},
	}

	_, err := r.client.ModifyKubernetesCluster(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify Kubernetes cluster",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	if !state.Version.Equal(plan.Version) {
		upgradeReq := request.UpgradeKubernetesClusterRequest{
			ClusterUUID: plan.ID.ValueString(),
			Upgrade: upcloud.KubernetesClusterUpgrade{
				Version: plan.Version.ValueString(),
			},
		}

		strategyType := plan.UpgradeStrategyType.ValueString()
		if strategyType != "" {
			upgradeReq.Upgrade.Strategy = &upcloud.KubernetesClusterUpgradeStrategy{
				Type: upcloud.KubernetesUpgradeStrategy(strategyType),
			}
		}

		_, err := r.client.UpgradeKubernetesCluster(ctx, &upgradeReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to upgrade Kubernetes cluster",
				utils.ErrorDiagnosticDetail(err),
			)
			return
		}
	}

	cluster, err := r.client.WaitForKubernetesClusterState(ctx, &request.WaitForKubernetesClusterStateRequest{
		DesiredState: upcloud.KubernetesClusterStateRunning,
		UUID:         plan.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for Kubernetes cluster to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setClusterValues(ctx, &plan, cluster)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kubernetesClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data kubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteKubernetesCluster(ctx, &request.DeleteKubernetesClusterRequest{
		UUID: data.ID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Kubernetes cluster",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	resp.Diagnostics.Append(waitForClusterToBeDeleted(ctx, r.client, data.ID.ValueString())...)
}

func (r *kubernetesClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
