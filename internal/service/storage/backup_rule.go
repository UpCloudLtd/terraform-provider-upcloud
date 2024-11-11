package storage

import (
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Lines > 1 should have one level of indentation to keep them under the right list item
const (
	BackupRuleSimpleBackupWarning = `Also, due to how UpCloud API works with simple backups and how Terraform orders the update operations, it is advised to never switch between ` + "`" + `simple_backup` + "`" + ` on the server and individual storages ` + "`" + `backup_rules` + "`" + ` in one apply. If you want to switch from using server simple backup to per-storage defined backup rules,  please first remove ` + "`" + `simple_backup` + "`" + ` block from a server, run ` + "`" + `terraform apply` + "`" + `, then add ` + "`" + `backup_rule` + "`" + ` to desired storages and run ` + "`" + `terraform apply` + "`" + ` again.`
	BackupRuleDescription         = `The criteria to backup the storage.

    Please keep in mind that it's not possible to have a storage with ` + "`" + `backup_rule` + "`" + ` attached to a server with ` + "`" + `simple_backup` + "`" + ` specified. Such configurations will throw errors during execution.

    ` + BackupRuleSimpleBackupWarning
)

type BackupRuleModel struct {
	Interval  types.String `tfsdk:"interval"`
	Time      types.String `tfsdk:"time"`
	Retention types.Int64  `tfsdk:"retention"`
}

func BackupRuleBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: BackupRuleDescription,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interval": schema.StringAttribute{
					Description: "The weekday when the backup is created",
					Required:    true,
				},
				"time": schema.StringAttribute{
					Description: "The time of day when the backup is created",
					Required:    true,
				},
				"retention": schema.Int64Attribute{
					Description: "The number of days before a backup is automatically deleted",
					Required:    true,
				},
			},
		},
	}
}

func BackupRuleSchema() *sdkv2_schema.Schema {
	return &sdkv2_schema.Schema{
		Description: BackupRuleDescription,
		Type:        sdkv2_schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &sdkv2_schema.Resource{
			Schema: map[string]*sdkv2_schema.Schema{
				"interval": {
					Description: "The weekday when the backup is created",
					Type:        sdkv2_schema.TypeString,
					Required:    true,
				},
				"time": {
					Description: "The time of day when the backup is created",
					Type:        sdkv2_schema.TypeString,
					Required:    true,
				},
				"retention": {
					Description: "The number of days before a backup is automatically deleted",
					Type:        sdkv2_schema.TypeInt,
					Required:    true,
				},
			},
		},
	}
}

func BackupRule(backupRule map[string]interface{}) *upcloud.BackupRule {
	if interval, ok := backupRule["interval"]; ok {
		if time, ok := backupRule["time"]; ok {
			if retention, ok := backupRule["retention"]; ok {
				return &upcloud.BackupRule{
					Interval:  interval.(string),
					Time:      time.(string),
					Retention: retention.(int),
				}
			}
		}
	}
	return &upcloud.BackupRule{}
}
