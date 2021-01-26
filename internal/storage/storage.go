package storage

import (
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func BackupRuleSchema() *schema.Schema {
	return &schema.Schema{
		Description: "The criteria to backup the storage",
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
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
