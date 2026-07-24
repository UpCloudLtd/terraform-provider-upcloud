package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerDefaultTitle(t *testing.T) {
	longHostname := strings.Repeat("x", 255)
	suffixLength := 24
	want := fmt.Sprintf("%s… (managed by terraform)", longHostname[0:255-suffixLength])
	got := defaultTitleFromHostname(longHostname)
	if want != got {
		t.Errorf("defaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}

	want = "terraform (managed by terraform)"
	got = defaultTitleFromHostname("terraform")
	if want != got {
		t.Errorf("defaultTitleFromHostname failed want '%s' got '%s'", want, got)
	}
}

func TestBuildSimpleBackupOpts_basic(t *testing.T) {
	value := simpleBackupModel{
		Time: types.StringValue("2200"),
		Plan: types.StringValue("weeklies"),
	}

	sb := buildSimpleBackupOpts(&value)
	expected := "2200,weeklies"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}

func TestBuildSimpleBackupOpts_withInvalidInput(t *testing.T) {
	value := simpleBackupModel{
		Time: types.StringValue("2200"),
	}

	sb := buildSimpleBackupOpts(&value)
	expected := "no"

	if sb != expected {
		t.Logf("BuildSimpleBackuOpts produced unexpected value. Expected: %s, received: %s", expected, sb)
		t.Fail()
	}
}

func newServerModelForTest(mutate func(*serverModel)) serverModel {
	m := serverModel{
		serverCommonModel: serverCommonModel{
			Timezone:          types.StringValue("UTC"),
			VideoModel:        types.StringValue("vga"),
			NICModel:          types.StringValue("virtio"),
			NetworkInterfaces: types.ListNull(types.ObjectType{}),
		},
		Template:       types.ListNull(types.ObjectType{}),
		StorageDevices: types.SetNull(types.ObjectType{}),
	}
	mutate(&m)
	return m
}

func TestChangeRequiresServerStop_withHotResize(t *testing.T) {
	tests := []struct {
		name           string
		state          serverModel
		plan           serverModel
		expectShutdown bool
	}{
		{
			name: "Only plan changing with hot_resize disabled",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(false)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("2xCPU-2GB")
				m.HotResize = types.BoolValue(true)
			}),
			expectShutdown: true,
		},
		{
			name: "Only plan changing with hot_resize enabled",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(true)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("2xCPU-2GB")
				m.HotResize = types.BoolValue(true)
			}),
			expectShutdown: false,
		},
		{
			name: "Custom plan with CPU changing and hot_resize enabled",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue(customPlanName)
				m.CPU = types.Int64Value(1)
				m.Mem = types.Int64Value(1024)
				m.HotResize = types.BoolValue(true)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue(customPlanName)
				m.CPU = types.Int64Value(2)
				m.Mem = types.Int64Value(1024)
				m.HotResize = types.BoolValue(true)
			}),
			expectShutdown: true,
		},
		{
			name: "Changing from standard plan to custom plan with different CPU/Mem and hot_resize enabled",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.CPU = types.Int64Value(1)
				m.Mem = types.Int64Value(1024)
				m.HotResize = types.BoolValue(true)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue(customPlanName)
				m.CPU = types.Int64Value(2)
				m.Mem = types.Int64Value(2048)
				m.HotResize = types.BoolValue(true)
			}),
			expectShutdown: true,
		},
		{
			name: "Only hot_resize changing from false to true",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(false)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(true)
			}),
			expectShutdown: false,
		},
		{
			name: "Only hot_resize changing from true to false",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(true)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(false)
			}),
			expectShutdown: false,
		},
		{
			name: "Hot resize with plan change and network change - should require shutdown",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(true)
				m.NetworkInterfaces = types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{},
							map[string]attr.Value{},
						),
					},
				)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("2xCPU-2GB")
				m.HotResize = types.BoolValue(true)
				m.NetworkInterfaces = types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{},
							map[string]attr.Value{},
						),
						types.ObjectValueMust(
							map[string]attr.Type{},
							map[string]attr.Value{},
						),
					},
				)
			}),
			expectShutdown: true,
		},
		{
			name: "Hot resize with plan change and storage change - should require shutdown",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(true)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("2xCPU-2GB")
				m.HotResize = types.BoolValue(true)
				m.Template = types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"size": types.Int64Type,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"size": types.Int64Type,
							},
							map[string]attr.Value{
								"size": types.Int64Value(20),
							},
						),
					},
				)
			}),
			expectShutdown: true,
		},
		{
			name: "Hot resize with plan change and template change - should require shutdown",
			state: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("1xCPU-1GB")
				m.HotResize = types.BoolValue(true)
			}),
			plan: newServerModelForTest(func(m *serverModel) {
				m.Plan = types.StringValue("2xCPU-2GB")
				m.HotResize = types.BoolValue(true)
				m.Template = types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{},
							map[string]attr.Value{},
						),
					},
				)
			}),
			expectShutdown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stateDevices, planDevices []storageDeviceModel
			tt.state.StorageDevices.ElementsAs(t.Context(), &stateDevices, false)
			tt.plan.StorageDevices.ElementsAs(t.Context(), &planDevices, false)
			result := changeRequiresServerStop(tt.state, tt.plan, stateDevices, planDevices)
			if result != tt.expectShutdown {
				t.Errorf("changeRequiresServerStop() = %v, want %v", result, tt.expectShutdown)
			}
		})
	}
}
