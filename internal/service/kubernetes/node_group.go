package kubernetes

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

var validTaintKeyRegExp = regexp.MustCompile("^[ -^`-~]+[ -~]*$") // Printable ASCII characters: ' ' (Space), ..., '^', '_', '`', ..., `~`

const (
	invalidTaintKeyMessage = "must only contain printable ASCII characters and must not start with an underscore"
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
	GPUPlan              types.List   `tfsdk:"gpu_plan"`
	CloudNativePlan      types.List   `tfsdk:"cloud_native_plan"`
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

type gpuPlanModel struct {
	StorageSize types.Int64  `tfsdk:"storage_size"`
	StorageTier types.String `tfsdk:"storage_tier"`
}

type cloudNativePlanModel struct {
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
			"labels": utils.LabelsAttributeWithValidators(
				"node_group",
				[]validator.String{stringvalidator.RegexMatches(utils.ValidLabelKeyRegExp, utils.InvalidLabelKeyMessage)},
				[]validator.String{stringvalidator.LengthBetween(0, 255)},
				mapplanmodifier.RequiresReplace(),
			),
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
				MarkdownDescription: "The storage encryption strategy to use for the nodes in this group. If not set, the cluster's storage encryption strategy will be used, if applicable. Valid values are `data-at-rest` and `none`.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				MarkdownDescription: "Resource properties for custom plan. This block is required for `custom` plans only.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					getCustomPlanPlanModifier(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cores": schema.Int64Attribute{
							MarkdownDescription: "The number of CPU cores dedicated to individual node group nodes.",
							Required:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(1, 20),
							},
						},
						"memory": schema.Int64Attribute{
							MarkdownDescription: "The amount of memory in megabytes to assign to individual node group node. Value needs to be divisible by 1024.",
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
								int64validator.Between(25, 4096),
							},
						},
						"storage_tier": schema.StringAttribute{
							MarkdownDescription: "The storage tier to use.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(string(upcloud.KubernetesStorageTierMaxIOPS)),
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
			"gpu_plan": schema.ListNestedBlock{
				MarkdownDescription: "Resource properties for GPU plan storage configuration. This block is optional for GPU plans.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					getGPUPlanPlanModifier(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"storage_size": schema.Int64Attribute{
							MarkdownDescription: "The size of the storage device in gigabytes.",
							Optional:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(25, 4096),
							},
						},
						"storage_tier": schema.StringAttribute{
							MarkdownDescription: "The storage tier to use.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(string(upcloud.KubernetesStorageTierMaxIOPS)),
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
			"cloud_native_plan": schema.ListNestedBlock{
				MarkdownDescription: "Resource properties for Cloud Native plan storage configuration. This block is optional for Cloud Native plans.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					getCloudNativePlanPlanModifier(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"storage_size": schema.Int64Attribute{
							MarkdownDescription: "The size of the storage device in gigabytes.",
							Optional:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(25, 4096),
							},
						},
						"storage_tier": schema.StringAttribute{
							MarkdownDescription: "The storage tier to use.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(string(upcloud.KubernetesStorageTierMaxIOPS)),
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
				MarkdownDescription: `Additional arguments for kubelet for the nodes in this group. Configure the arguments without leading ` + "`" + `--` + "`" + `. The API will prefix the arguments with ` + "`" + `--` + "`" + ` when preparing kubelet call.

    Note that these arguments will be passed directly to kubelet CLI on each worker node without any validation. Passing invalid arguments can break your whole cluster. Be extra careful when adding kubelet args.`,
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
								stringvalidator.RegexMatches(validTaintKeyRegExp, invalidTaintKeyMessage),
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

	// Build the appropriate plan based on plan type
	var customPlan *upcloud.KubernetesNodeGroupCustomPlan
	var gpuPlan *upcloud.KubernetesNodeGroupGPUPlan
	var cloudNativePlan *upcloud.KubernetesNodeGroupCloudNativePlan
	var diags diag.Diagnostics

	planType := data.Plan.ValueString()
	if planType == "custom" {
		customPlan, diags = buildCustomPlan(ctx, data.CustomPlan)
		resp.Diagnostics.Append(diags...)
	} else if strings.HasPrefix(planType, gpuPlanPrefix) {
		gpuPlan, diags = buildGPUPlan(ctx, data.GPUPlan)
		resp.Diagnostics.Append(diags...)
	} else if strings.HasPrefix(planType, cloudNativePlanPrefix) {
		cloudNativePlan, diags = buildCloudNativePlan(ctx, data.CloudNativePlan)
		resp.Diagnostics.Append(diags...)
	}

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
			GPUPlan:              gpuPlan,
			CloudNativePlan:      cloudNativePlan,
			StorageEncryption:    upcloud.StorageEncryption(data.StorageEncryption.ValueString()),
		},
	}

	ng, err := r.client.CreateKubernetesNodeGroup(ctx, &apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Kubernetes node group",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.Cluster.ValueString(), ng.Name))

	ng, err = r.client.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
		ClusterUUID:  data.Cluster.ValueString(),
		Name:         ng.Name,
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

	if data.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)

		return
	}

	var cluster, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &cluster, &name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal Kubernetes node group name",
			utils.ErrorDiagnosticDetail(err),
		)

		return
	}

	nodeGroup, err := r.client.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: cluster,
		Name:        name,
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

	var nodeCountState types.Int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("node_count"), &nodeCountState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Compare node count attribute value between plan and prior state
	if data.NodeCount.Equal(nodeCountState) {
		return
	}

	apiReq := &request.ModifyKubernetesNodeGroupRequest{
		ClusterUUID: data.Cluster.ValueString(),
		Name:        data.Name.ValueString(),
		NodeGroup: request.ModifyKubernetesNodeGroup{
			Count: int(data.NodeCount.ValueInt64()),
		},
	}

	ng, err := r.client.ModifyKubernetesNodeGroup(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to modify Kubernetes node group",
			utils.ErrorDiagnosticDetail(err),
		)
		return
	}

	data.ID = types.StringValue(utils.MarshalID(data.Cluster.ValueString(), ng.Name))

	ng, err = r.client.WaitForKubernetesNodeGroupState(ctx, &request.WaitForKubernetesNodeGroupStateRequest{
		DesiredState: upcloud.KubernetesNodeGroupStateRunning,
		ClusterUUID:  data.Cluster.ValueString(),
		Name:         ng.Name,
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
}

func (r *kubernetesNodeGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setNodeGroupValues(ctx context.Context, data *kubernetesNodeGroupModel, ng *upcloud.KubernetesNodeGroup) diag.Diagnostics {
	var diags, respDiagnostics diag.Diagnostics

	isImport := data.Cluster.ValueString() == ""

	var cluster, name string
	err := utils.UnmarshalID(data.ID.ValueString(), &cluster, &name)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to unmarshal Kubernetes node group name",
			utils.ErrorDiagnosticDetail(err),
		)
	}

	data.AntiAffinity = types.BoolValue(ng.AntiAffinity)
	data.Cluster = types.StringValue(cluster)

	// Handle custom plan
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

	// Handle GPU plan
	if isImport || !data.GPUPlan.IsNull() {
		gpuPlans := make([]gpuPlanModel, 0)
		if ng.GPUPlan != nil {
			gpuPlans = append(gpuPlans, gpuPlanModel{
				StorageSize: types.Int64Value(int64(ng.GPUPlan.StorageSize)),
				StorageTier: types.StringValue(string(ng.GPUPlan.StorageTier)),
			})
		}
		data.GPUPlan, diags = types.ListValueFrom(ctx, data.GPUPlan.ElementType(ctx), gpuPlans)
		respDiagnostics.Append(diags...)
	}

	// Handle Cloud Native plan
	if isImport || !data.CloudNativePlan.IsNull() {
		cloudNativePlans := make([]cloudNativePlanModel, 0)
		if ng.CloudNativePlan != nil {
			cloudNativePlans = append(cloudNativePlans, cloudNativePlanModel{
				StorageSize: types.Int64Value(int64(ng.CloudNativePlan.StorageSize)),
				StorageTier: types.StringValue(string(ng.CloudNativePlan.StorageTier)),
			})
		}
		data.CloudNativePlan, diags = types.ListValueFrom(ctx, data.CloudNativePlan.ElementType(ctx), cloudNativePlans)
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

	data.Name = types.StringValue(ng.Name)
	data.NodeCount = types.Int64Value(int64(ng.Count))
	data.Plan = types.StringValue(ng.Plan)

	if isImport || !data.SSHKeys.IsNull() {
		data.SSHKeys, diags = types.SetValueFrom(ctx, types.StringType, ng.SSHKeys)
		respDiagnostics.Append(diags...)
	}

	data.StorageEncryption = types.StringValue(string(ng.StorageEncryption))

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
		customPlan := upcloud.KubernetesNodeGroupCustomPlan{}

		// Cores and Memory are only set for custom plans, not for GPU or Cloud Native plans
		if !plan.Cores.IsNull() && !plan.Cores.IsUnknown() {
			customPlan.Cores = int(plan.Cores.ValueInt64())
		}
		if !plan.Memory.IsNull() && !plan.Memory.IsUnknown() {
			customPlan.Memory = int(plan.Memory.ValueInt64())
		}

		// StorageSize is optional for GPU and Cloud Native plans
		if !plan.StorageSize.IsNull() && !plan.StorageSize.IsUnknown() {
			customPlan.StorageSize = int(plan.StorageSize.ValueInt64())
		}

		// StorageTier is optional for all plan types
		if !plan.StorageTier.IsNull() && !plan.StorageTier.IsUnknown() {
			customPlan.StorageTier = upcloud.StorageTier(plan.StorageTier.ValueString())
		}

		customPlans = append(customPlans, customPlan)
	}

	return &customPlans[0], respDiagnostics
}

func buildGPUPlan(ctx context.Context, dataGPUPlans types.List) (*upcloud.KubernetesNodeGroupGPUPlan, diag.Diagnostics) {
	var planGPUPlans []gpuPlanModel
	respDiagnostics := dataGPUPlans.ElementsAs(ctx, &planGPUPlans, false)

	if len(planGPUPlans) == 0 {
		return nil, respDiagnostics
	}

	gpuPlans := make([]upcloud.KubernetesNodeGroupGPUPlan, 0)
	for _, plan := range planGPUPlans {
		gpuPlan := upcloud.KubernetesNodeGroupGPUPlan{}

		// For GPU plans, only storage configuration is allowed
		if !plan.StorageSize.IsNull() && !plan.StorageSize.IsUnknown() {
			gpuPlan.StorageSize = int(plan.StorageSize.ValueInt64())
		}

		if !plan.StorageTier.IsNull() && !plan.StorageTier.IsUnknown() {
			gpuPlan.StorageTier = upcloud.StorageTier(plan.StorageTier.ValueString())
		}

		gpuPlans = append(gpuPlans, gpuPlan)
	}

	return &gpuPlans[0], respDiagnostics
}

func buildCloudNativePlan(ctx context.Context, dataCloudNativePlans types.List) (*upcloud.KubernetesNodeGroupCloudNativePlan, diag.Diagnostics) {
	var planCloudNativePlans []cloudNativePlanModel
	respDiagnostics := dataCloudNativePlans.ElementsAs(ctx, &planCloudNativePlans, false)

	if len(planCloudNativePlans) == 0 {
		return nil, respDiagnostics
	}

	cloudNativePlans := make([]upcloud.KubernetesNodeGroupCloudNativePlan, 0)
	for _, plan := range planCloudNativePlans {
		cloudNativePlan := upcloud.KubernetesNodeGroupCloudNativePlan{}

		// For Cloud Native plans, only storage configuration is allowed
		if !plan.StorageSize.IsNull() && !plan.StorageSize.IsUnknown() {
			cloudNativePlan.StorageSize = int(plan.StorageSize.ValueInt64())
		}

		if !plan.StorageTier.IsNull() && !plan.StorageTier.IsUnknown() {
			cloudNativePlan.StorageTier = upcloud.StorageTier(plan.StorageTier.ValueString())
		}

		cloudNativePlans = append(cloudNativePlans, cloudNativePlan)
	}

	return &cloudNativePlans[0], respDiagnostics
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
	return diags
}
