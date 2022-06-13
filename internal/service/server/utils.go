package server

import (
	"context"
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func isProviderAccountSubaccount(ctx context.Context, s *service.ServiceContext) (bool, error) {
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

func hasTemplateBackupRuleBeenReplacedWithSimpleBackups(d *schema.ResourceData) bool {
	if !d.HasChange("simple_backup") || !d.HasChange("template.0.backup_rule") {
		return false
	}

	sb, sbOk := d.GetOk("simple_backup")
	if !sbOk {
		return false
	}

	simpleBackup := sb.(*schema.Set).List()[0].(map[string]interface{})
	if simpleBackup["interval"] == "" {
		return false
	}

	tbr, tbrOk := d.GetOk("template.0.backup_rule.0")
	templateBackupRule := tbr.(map[string]interface{})
	if tbrOk && templateBackupRule["interval"] != "" {
		return false
	}

	return true
}

func sliceToMap(input []string) map[string]bool {
	output := make(map[string]bool)
	for _, i := range input {
		output[i] = true
	}
	return output
}
