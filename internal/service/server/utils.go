package server

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func isProviderAccountSubaccount(ctx context.Context, s *service.Service) (bool, error) {
	account, err := s.GetAccount(ctx)
	if err != nil {
		return false, err
	}
	a, err := s.GetAccountDetails(ctx, &request.GetAccountDetailsRequest{Username: account.UserName})
	if err != nil {
		return false, err
	}
	return a.IsSubaccount(), nil
}

func defaultTitleFromHostname(hostname string) string {
	const suffix string = " (managed by terraform)"
	if len(hostname)+len(suffix) > serverTitleLength {
		hostname = fmt.Sprintf("%s…", hostname[:serverTitleLength-len(suffix)-1])
	}
	return fmt.Sprintf("%s%s", hostname, suffix)
}

func hasTemplateBackupRuleBeenReplacedWithSimpleBackups(ctx context.Context, state, plan serverModel) (yes bool, diags diag.Diagnostics) {
	stateTemplate, d := getTemplate(ctx, state)
	diags.Append(d...)

	planTemplate, d := getTemplate(ctx, plan)
	diags.Append(d...)

	if diags.HasError() {
		return
	}

	if plan.SimpleBackup.Equal(state.SimpleBackup) || planTemplate.BackupRule.Equal(stateTemplate.BackupRule) {
		return
	}

	if plan.SimpleBackup.IsNull() {
		return
	}

	if planTemplate.BackupRule.IsNull() {
		yes = true
	}

	return
}

func sliceToMap(input []string) map[string]bool {
	output := make(map[string]bool)
	for _, i := range input {
		output[i] = true
	}
	return output
}

func changeRequiresServerStop(state, plan serverModel) bool {
	return !plan.CPU.Equal(state.CPU) ||
		!plan.Mem.Equal(state.Mem) ||
		!plan.Plan.Equal(state.Plan) ||
		!plan.Timezone.Equal(state.Timezone) ||
		!plan.VideoModel.Equal(state.VideoModel) ||
		!plan.NICModel.Equal(state.NICModel) ||
		!plan.Template.Equal(state.Template) ||
		!plan.StorageDevices.Equal(state.StorageDevices) ||
		!plan.NetworkInterfaces.Equal(state.NetworkInterfaces)
}
