package storage

import (
	"github.com/UpCloudLtd/upcloud-go-api/v4/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func BackupRuleSchema() *schema.Schema {
	return &schema.Schema{
		Description: `The criteria to backup the storage  
		Please keep in mind that it's not possible to have a server with backup_rule attached to a server with simple_backup specified.
		Such configurations will throw errors during execution.  
		Also, due to how UpCloud API works with simple backups and how Terraform orders the update operations, 
		it is advised to never switch between simple_backup on the server and individual storages backup_rules in one apply.
		If you want to switch from using server simple backup to per-storage defined backup rules, 
		please first remove simple_backup block from a server, run 'terraform apply', 
		then add 'backup_rule' to desired storages and run 'terraform apply' again.`,
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"interval": {
					Description: "The weekday when the backup is created",
					Type:        schema.TypeString,
					Required:    true,
				},
				"time": {
					Description: "The time of day when the backup is created",
					Type:        schema.TypeString,
					Required:    true,
				},
				"retention": {
					Description: "The number of days before a backup is automatically deleted",
					Type:        schema.TypeInt,
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
