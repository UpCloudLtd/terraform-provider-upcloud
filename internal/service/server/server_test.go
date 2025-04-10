package server

import (
	"fmt"
	"strings"
	"testing"

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

func TestChangeRequiresServerStop_withHotResize(t *testing.T) {
	// Define default values for fields not being tested to ensure they're equal
	defaultTimezone := types.StringValue("UTC")
	defaultVideoModel := types.StringValue("vga")
	defaultNICModel := types.StringValue("virtio")
	defaultTemplate := types.ListNull(types.ObjectType{})
	defaultStorageDevices := types.SetNull(types.ObjectType{})
	defaultNetworkInterfaces := types.ListNull(types.ObjectType{})

	tests := []struct {
		name           string
		state          serverModel
		plan           serverModel
		expectShutdown bool
	}{
		{
			name: "Only plan changing with hot_resize disabled",
			state: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				HotResize:         types.BoolValue(false),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			plan: serverModel{
				Plan:              types.StringValue("2xCPU-2GB"),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			expectShutdown: true,
		},
		{
			name: "Only plan changing with hot_resize enabled",
			state: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			plan: serverModel{
				Plan:              types.StringValue("2xCPU-2GB"),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			expectShutdown: false,
		},
		{
			name: "Custom plan with CPU changing and hot_resize enabled",
			state: serverModel{
				Plan:              types.StringValue(customPlanName),
				CPU:               types.Int64Value(1),
				Mem:               types.Int64Value(1024),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			plan: serverModel{
				Plan:              types.StringValue(customPlanName),
				CPU:               types.Int64Value(2),
				Mem:               types.Int64Value(1024),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			expectShutdown: true,
		},
		{
			name: "Changing from standard plan to custom plan with different CPU/Mem and hot_resize enabled",
			state: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				CPU:               types.Int64Value(1),
				Mem:               types.Int64Value(1024),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			plan: serverModel{
				Plan:              types.StringValue(customPlanName),
				CPU:               types.Int64Value(2),
				Mem:               types.Int64Value(2048),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			expectShutdown: true,
		},
		{
			name: "Only hot_resize changing from false to true",
			state: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				HotResize:         types.BoolValue(false),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			plan: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			expectShutdown: false,
		},
		{
			name: "Only hot_resize changing from true to false",
			state: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				HotResize:         types.BoolValue(true),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			plan: serverModel{
				Plan:              types.StringValue("1xCPU-1GB"),
				HotResize:         types.BoolValue(false),
				Timezone:          defaultTimezone,
				VideoModel:        defaultVideoModel,
				NICModel:          defaultNICModel,
				Template:          defaultTemplate,
				StorageDevices:    defaultStorageDevices,
				NetworkInterfaces: defaultNetworkInterfaces,
			},
			expectShutdown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := changeRequiresServerStop(tt.state, tt.plan)
			if result != tt.expectShutdown {
				t.Errorf("changeRequiresServerStop() = %v, want %v", result, tt.expectShutdown)
			}
		})
	}
}
