package kubernetes

import (
	"context"
	"fmt"
	"regexp"
	"time"

	planmodifierutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/planmodifier"
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	validatorutil "github.com/UpCloudLtd/terraform-provider-upcloud/internal/validator"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &kubernetesNodeGroupResource{}
	_ resource.ResourceWithConfigure   = &kubernetesNodeGroupResource{}
	_ resource.ResourceWithImportState = &kubernetesNodeGroupResource{}
)

func NewKubernetesNodeGroupResource() resource.Resource {
	return &kubernetesNodeGroupResource{}
}

type kubernetesNodeGroupResource struct {
	client *service.Service
}

func (r *kubernetesNodeGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_node_group"
}

// Configure adds the provider configured client to the resource.
func (r *kubernetesNodeGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client, resp.Diagnostics = utils.GetClientFromProviderData(req.ProviderData)
}

type kubernetesNodeGroupModel struct {
	AntiAffinity         types.Bool   `tfsdk:"anti_affinity"`
	Cluster              types.String `tfsdk:"cluster"`
	CustomPlan           types.List   `tfsdk:"custom_plan"`
	KubeletArgs          types.Set    `tfsdk:"kubelet_args"`
	ID                   types.String `tfsdk:"id"`
	Labels               types.Map    `tfsdk:"labels"`
	Name                 types.String `tfsdk:"name"`
	NodeCount            types.Int64  `tfsdk:"node_count"`
	Plan                 types.String `tfsdk:"plan"`
	SSHKeys              types.Set    `tfsdk:"ssh_keys"`
	StorageEncryption    types.String `tfsdk:"storage_encryption"`
	Taint                types.Set    `tfsdk:"taint"`
	UtilityNetworkAccess types.Bool   `tfsdk:"utility_network_access"`
}

type customPlanModel struct {
	Cores       types.Int64  `tfsdk:"cores"`
	Memory      types.Int64  `tfsdk:"memory"`
	StorageSize types.Int64  `tfsdk:"storage_size"`
	StorageTier types.String `tfsdk:"storage_tier"`
}

type kubeletArgModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type taintModel struct {
	Effect types.String `tfsdk:"effect"`
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
}

func (r *kubernetesNodeGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource represents a [Managed Kubernetes](https://upcloud.com/products/managed-kubernetes) cluster.",
		Attributes: map[string]schema.Attribute{
			"anti_affinity": schema.BoolAttribute{
				MarkdownDescription: "If set to true, nodes in this group will be placed on separate compute hosts. Please note that anti-affinity policy is considered 'best effort' and enabling it does not fully guarantee that the nodes will end up on different hardware.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"cluster": schema.StringAttribute{
				MarkdownDescription: idDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Computed ID of the node group. This is a combination of the cluster UUID and the node group name, separated with a `/`.",
				Computed:            true,
			},
			"labels": utils.LabelsAttribute("node_group", mapplanmodifier.RequiresReplace()),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the node group. Needs to be unique within a cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(resourceNameMaxLength),
					stringvalidator.RegexMatches(resourceNameRegexp, fmt.Sprintf("name should only contain lowercase alphanumeric characters and dashes (a-z, 0-9, -). Name should not start or end with a dash. Regular expresion used to check validation: %s", resourceNameRegexp)),
				},
			},
			"node_count": schema.Int64Attribute{
				MarkdownDescription: "Amount of nodes to provision in the node group.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "The server plan used for the node group. You can list available plans with `upctl server plans`",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_keys": schema.SetAttribute{
				MarkdownDescription: "You can optionally select SSH keys to be added as authorized keys to the nodes in this node group. This allows you to connect to the nodes via SSH once they are running.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default: setdefault.StaticValue(
					types.SetValueMust(
						types.StringType,
						[]attr.Value{},
					),
				),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"storage_encryption": schema.StringAttribute{
				MarkdownDescription: "The storage encryption strategy to use for the nodes in this group. If not set, the cluster's storage encryption strategy will be used, if applicable.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(upcloud.StorageEncryptionDataAtRest),
						string(upcloud.StorageEncryptionNone),
					),
				},
			},
			"utility_network_access": schema.BoolAttribute{
				MarkdownDescription: "If set to false, nodes in this group will not have access to utility network.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"custom_plan": schema.ListNestedBlock{
				MarkdownDescription: "Resource properties for custom plan",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					planmodifierutil.CustomPlanPlanModifier(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cores": schema.Int64Attribute{
							MarkdownDescription: "The number of CPU cores dedicated to individual node group nodes when using custom plan",
							Required:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(1, 20),
							},
						},
						"memory": schema.Int64Attribute{
							MarkdownDescription: "The amount of memory in megabytes to assign to individual node group node when using custom plan. Value needs to be divisible by 1024.",
							Required:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(2048, 131072),
								validatorutil.DivisibleBy(1024),
							},
						},
						"storage_size": schema.Int64Attribute{
							MarkdownDescription: "The size of the storage device in gigabytes.",
							Required:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(25, 1024),
							},
						},
						"storage_tier": schema.StringAttribute{
							MarkdownDescription: fmt.Sprintf("The storage tier to use. Defaults to %s", upcloud.KubernetesStorageTierMaxIOPS),
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(upcloud.KubernetesStorageTierMaxIOPS),
									string(upcloud.KubernetesStorageTierHDD),
									string(upcloud.KubernetesStorageTierStandard),
								),
							},
						},
					},
				},
			},
			"kubelet_args": schema.SetNestedBlock{
				MarkdownDescription: "Additional arguments for kubelet for the nodes in this group. WARNING - those arguments will be passed directly to kubelet CLI on each worker node without any validation. Passing invalid arguments can break your whole cluster. Be extra careful when adding kubelet args.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "Kubelet argument key.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9-]+$"), "needs to match regexp ^[a-zA-Z0-9-]+$"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Kubelet argument value.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 255),
							},
						},
					},
				},
			},
			"taint": schema.SetNestedBlock{
				MarkdownDescription: "Taints for the nodes in this group.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							MarkdownDescription: "Taint effect.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(upcloud.KubernetesClusterTaintEffectNoExecute),
									string(upcloud.KubernetesClusterTaintEffectNoSchedule),
									string(upcloud.KubernetesClusterTaintEffectPreferNoSchedule),
								),
							},
						},
						"key": schema.StringAttribute{
							MarkdownDescription: "Taint key.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9-]+$"), "needs to match regexp ^[a-zA-Z0-9-]+$"),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Taint value.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 255),
							},
						},
					},
				},
			},
		},
	}
}

func (r *kubernetesNodeGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	diags := r.modifyPlanStorageEncryption(ctx, req, resp)
	resp.Diagnostics.Append(diags...)
}

// modifyPlanStorageEncryption checks if cluster has storage encryption strategy set and applies that value to the node group when applicable.
// Purpose for this is to make storage_encryption attribute known *before* apply if it's not defined.
func (r *kubernetesNodeGroupResource) modifyPlanStorageEncryption(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) diag.Diagnostics {
	var storageEncryption types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("storage_encryption"), &storageEncryption)
	if diags.HasError() {
		return diags
	}

	if !storageEncryption.IsNull() || !storageEncryption.IsUnknown() {
		return diags
	}

	var clusterUUID types.String
	diags.Append(req.Plan.GetAttribute(ctx, path.Root("cluster"), &clusterUUID)...)
	if diags.HasError() {
		return diags
	}

	if clusterUUID.ValueString() == "" {
		return diags
	}

	c, err := r.client.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{UUID: clusterUUID.ValueString()})
	if err != nil {
		diags.AddError(
			"Unable to get Kubernetes cluster",
			utils.ErrorDiagnosticDetail(err),
		)

		return diags
	}

	if c.StorageEncryption == "" {
		return diags
	}

	resp.Plan.SetAttribute(
		ctx,
		path.Root("storage_encryption"),
		types.StringValue(string(c.StorageEncryption)),
	)

	return diags
}

func (r *kubernetesNodeGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data kubernetesNodeGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var labels map[string]string
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labels, false)...)
	}

	var sshKeys []string
	if !data.SSHKeys.IsNull() && !data.SSHKeys.IsUnknown() {
		resp.Diagnostics.Append(data.SSHKeys.ElementsAs(ctx, &sshKeys, false)...)
	}

	customPlan, diags := buildCustomPlan(ctx, data.CustomPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kubeletArgs, diags := buildKubeletArgs(ctx, data.KubeletArgs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	taints, diags := buildTaints(ctx, data.Taint)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := request.CreateKubernetesNodeGroupRequest{
		ClusterUUID: data.Cluster.ValueString(),
		NodeGroup: request.KubernetesNodeGroup{
			Count:                int(data.NodeCount.ValueInt64()),
			Labels:               utils.LabelsMapToSlice(labels),
			Name:                 data.Name.ValueString(),
			Plan:                 data.Plan.ValueString(),
			SSHKeys:              sshKeys,
			Storage:              "",
			KubeletArgs:          kubeletArgs,
			Taints:               taints,
			AntiAffinity:         data.AntiAffinity.ValueBool(),
			UtilityNetworkAccess: data.UtilityNetworkAccess.ValueBoolPointer(),
			CustomPlan:           customPlan,
			StorageEncryption:    upcloud.StorageEncryption(data.StorageEncryption.ValueString()),
		},
	}

	_, err := r.client.CreateKubernetesNodeGroup(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Kubernetes node group",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	ng, err := r.client.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
		ClusterUUID:  data.Cluster.ValueString(),
		Name:         data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for Kubernetes node group to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setNodeGroupValues(ctx, &data, ng)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesNodeGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kubernetesNodeGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	nodeGroup, err := r.client.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: data.Cluster.ValueString(),
		Name:        data.Name.ValueString(),
	})
	if err != nil {
		if utils.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to read Kubernetes node group details",
				utils.ErrorDiagnosticDetail(err),
			)
		}
		return
	}

	resp.Diagnostics.Append(setNodeGroupValues(ctx, &data, &nodeGroup.KubernetesNodeGroup)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesNodeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data kubernetesNodeGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var nodeCountPlan, nodeCountState types.Int64
	var clusterUUID, name types.String

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("node_count"), &nodeCountPlan)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("cluster"), &clusterUUID)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &name)...)

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("node_count"), &nodeCountState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Compare node count attribute value between plan and prior state
	if nodeCountPlan.Equal(nodeCountState) {
		return
	}

	apiReq := &request.ModifyKubernetesNodeGroupRequest{
		ClusterUUID: clusterUUID.ValueString(),
		Name:        name.ValueString(),
		NodeGroup: request.ModifyKubernetesNodeGroup{
			Count: int(nodeCountPlan.ValueInt64()),
		},
	}

	_, err := r.client.ModifyKubernetesNodeGroup(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify Kubernetes node group",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	ng, err := r.client.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
		ClusterUUID:  apiReq.ClusterUUID,
		Name:         apiReq.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while waiting for Kubernetes ng to be in running state",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(setNodeGroupValues(ctx, &data, ng)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesNodeGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data kubernetesNodeGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if err := r.client.DeleteKubernetesNodeGroup(ctx, &request.DeleteKubernetesNodeGroupRequest{
		ClusterUUID: data.Cluster.ValueString(),
		Name:        data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Kubernetes node group",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	// wait before continuing so that all nodes are destroyed
	resp.Diagnostics.Append(waitForNodeGroupToBeDeleted(ctx, r.client, data.Cluster.ValueString(), data.Name.ValueString())...)

	// If there was an error during while waiting for the node group to be deleted - just end the delete operation here
	if resp.Diagnostics.HasError() {
		return
	}

	// Additionally wait some time so that all cleanup operations can finish
	time.Sleep(time.Second * cleanupWaitTimeSeconds)
}

func (r *kubernetesNodeGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var cluster, name string
	err := utils.UnmarshalID(req.ID, &cluster, &name)

	if err != nil || cluster == "" || name == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: cluster/name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), cluster)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func setNodeGroupValues(ctx context.Context, data *kubernetesNodeGroupModel, ng *upcloud.KubernetesNodeGroup) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.ID.ValueString() == ""

	data.AntiAffinity = types.BoolValue(ng.AntiAffinity)
	data.Cluster = types.StringValue(data.Cluster.ValueString())

	if isImport || !data.CustomPlan.IsNull() {
		customPlans := make([]customPlanModel, 0)
		if ng.CustomPlan != nil {
			customPlans = append(customPlans, customPlanModel{
				Cores:       types.Int64Value(int64(ng.CustomPlan.Cores)),
				Memory:      types.Int64Value(int64(ng.CustomPlan.Memory)),
				StorageSize: types.Int64Value(int64(ng.CustomPlan.StorageSize)),
				StorageTier: types.StringValue(string(ng.CustomPlan.StorageTier)),
			})
		}

		data.CustomPlan, diags = types.ListValueFrom(ctx, data.CustomPlan.ElementType(ctx), customPlans)
		respDiagnostics.Append(diags...)
	}

	if isImport || !data.KubeletArgs.IsNull() {
		kubeletArgs := make([]kubeletArgModel, 0)
		for _, arg := range ng.KubeletArgs {
			kubeletArgs = append(kubeletArgs, kubeletArgModel{
				Key:   types.StringValue(arg.Key),
				Value: types.StringValue(arg.Value),
			})
		}

		data.KubeletArgs, diags = types.SetValueFrom(ctx, data.KubeletArgs.ElementType(ctx), kubeletArgs)
		respDiagnostics.Append(diags...)
	}

	data.Labels, diags = types.MapValueFrom(ctx, types.StringType, utils.LabelsSliceToMap(ng.Labels))
	respDiagnostics.Append(diags...)

	data.ID = types.StringValue(utils.MarshalID(data.Cluster.ValueString(), ng.Name))
	data.Name = types.StringValue(ng.Name)
	data.NodeCount = types.Int64Value(int64(ng.Count))
	data.Plan = types.StringValue(ng.Plan)

	if isImport || !data.SSHKeys.IsNull() {
		data.SSHKeys, diags = types.SetValueFrom(ctx, types.StringType, ng.SSHKeys)
		respDiagnostics.Append(diags...)
	}

	if isImport || !data.StorageEncryption.IsNull() {
		data.StorageEncryption = types.StringValue(string(ng.StorageEncryption))
	}

	if isImport || !data.Taint.IsNull() {
		taints := make([]taintModel, 0)
		for _, taint := range ng.Taints {
			taints = append(taints, taintModel{
				Effect: types.StringValue(string(taint.Effect)),
				Key:    types.StringValue(taint.Key),
				Value:  types.StringValue(taint.Value),
			})
		}

		data.Taint, diags = types.SetValueFrom(ctx, data.Taint.ElementType(ctx), taints)
		respDiagnostics.Append(diags...)
	}

	data.UtilityNetworkAccess = types.BoolValue(ng.UtilityNetworkAccess)

	return respDiagnostics
}

func buildCustomPlan(ctx context.Context, dataCustomPlans types.List) (*upcloud.KubernetesNodeGroupCustomPlan, diag.Diagnostics) {
	var planCustomPlans []customPlanModel
	respDiagnostics := dataCustomPlans.ElementsAs(ctx, &planCustomPlans, false)

	if len(planCustomPlans) == 0 {
		return nil, respDiagnostics
	}

	customPlans := make([]upcloud.KubernetesNodeGroupCustomPlan, 0)
	for _, plan := range planCustomPlans {
		customPlan := upcloud.KubernetesNodeGroupCustomPlan{
			Cores:       int(plan.Cores.ValueInt64()),
			Memory:      int(plan.Memory.ValueInt64()),
			StorageSize: int(plan.StorageSize.ValueInt64()),
		}
		if !plan.StorageTier.IsNull() || !plan.StorageTier.IsUnknown() {
			customPlan.StorageTier = upcloud.StorageTier(plan.StorageTier.ValueString())
		}

		customPlans = append(customPlans, customPlan)
	}

	return &customPlans[0], respDiagnostics
}

func buildKubeletArgs(ctx context.Context, dataKubeletArgs types.Set) ([]upcloud.KubernetesKubeletArg, diag.Diagnostics) {
	var planKubeletArgs []kubeletArgModel
	respDiagnostics := dataKubeletArgs.ElementsAs(ctx, &planKubeletArgs, false)
	kubeletArgs := make([]upcloud.KubernetesKubeletArg, 0)

	for _, kubeletArg := range planKubeletArgs {
		kubeletArgs = append(kubeletArgs, upcloud.KubernetesKubeletArg{
			Key:   kubeletArg.Key.ValueString(),
			Value: kubeletArg.Value.ValueString(),
		})
	}

	return kubeletArgs, respDiagnostics
}

func buildTaints(ctx context.Context, dataTaints types.Set) ([]upcloud.KubernetesTaint, diag.Diagnostics) {
	var planTaints []taintModel
	respDiagnostics := dataTaints.ElementsAs(ctx, &planTaints, false)
	taints := make([]upcloud.KubernetesTaint, 0)

	for _, taint := range planTaints {
		taints = append(taints, upcloud.KubernetesTaint{
			Effect: upcloud.KubernetesClusterTaintEffect(taint.Effect.ValueString()),
			Key:    taint.Key.ValueString(),
			Value:  taint.Value.ValueString(),
		})
	}

	return taints, respDiagnostics
}

func getNodeGroupDeleted(ctx context.Context, svc *service.Service, id ...string) (map[string]interface{}, error) {
	c, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: id[0],
		Name:        id[1],
	})

	return map[string]interface{}{"resource": "node-group", "name": c.Name, "state": c.State}, err
}

func waitForNodeGroupToBeDeleted(ctx context.Context, svc *service.Service, clusterUUID, name string) (diags diag.Diagnostics) {
	err := utils.WaitForResourceToBeDeleted(ctx, svc, getNodeGroupDeleted, clusterUUID, name)
	if err != nil {
		diags.AddError("Error waiting for node group to be deleted", err.Error())
	}
	return
}
