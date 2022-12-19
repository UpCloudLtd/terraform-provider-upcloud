package database

import (
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v5/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceRedis() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents Redis managed database",
		CreateContext: resourceDatabaseCreate(upcloud.ManagedDatabaseServiceTypeRedis),
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(),
			schemaRedisEngine(),
		),
	}
}

func schemaRedisEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"properties": {
			Description: "Database Engine properties for Redis",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: utils.JoinSchemas(
					schemaDatabaseCommonProperties(),
					schemaRedisProperties(),
				),
			},
		},
	}
}

func schemaRedisProperties() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redis_lfu_decay_time": {
			Type:        schema.TypeInt,
			Description: "LFU maxmemory-policy counter decay time in minutes. Default is 1.",
			Computed:    true,
			Optional:    true,
		},
		"redis_number_of_databases": {
			Type:        schema.TypeInt,
			Description: "Number of redis databases. Set number of redis databases. Changing this will cause a restart of redis service.",
			Optional:    true,
			Computed:    true,
		},
		"redis_notify_keyspace_events": {
			Type:        schema.TypeString,
			Description: "Set notify-keyspace-events option. Default is \"\".",
			Optional:    true,
			Default:     true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.All(
					validation.StringLenBetween(0, 32),
					validation.StringMatch(regexp.MustCompile(`^[KEg\\$lshzxeA]*$`), "must match '^[KEg\\$lshzxeA]*$' pattern")),
			),
		},
		"redis_pubsub_client_output_buffer_limit": {
			Type:        schema.TypeInt,
			Description: "Pub/sub client output buffer hard limit in MB. Set output buffer limit for pub / sub clients in MB. The value is the hard limit, the soft limit is 1/4 of the hard limit. When setting the limit, be mindful of the available memory in the selected service plan.",
			Optional:    true,
			Computed:    true,
		},
		"redis_ssl": {
			Type:        schema.TypeBool,
			Description: "Require SSL to access Redis. Default is `true`.",
			Default:     true,
			Optional:    true,
		},
		"recovery_basebackup_name": {
			Type:        schema.TypeString,
			Description: "Name of the basebackup to restore in forked service.",
			Optional:    true,
			Computed:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9-_:.]+$"), "must match '^[a-zA-Z0-9-_:.]+$' pattern e.g. 'backup-20191112t091354293891z'")),
			),
		},
		"redis_lfu_log_factor": {
			Type:        schema.TypeInt,
			Description: "Counter logarithm factor for volatile-lfu and allkeys-lfu maxmemory-policies. Default is 10.",
			Computed:    true,
			Optional:    true,
		},
		"redis_io_threads": {
			Type:        schema.TypeInt,
			Description: "Redis IO thread count.",
			Optional:    true,
			Computed:    true,
		},
		"redis_maxmemory_policy": {
			Type:        schema.TypeString,
			Description: "Redis maxmemory-policy. Default is `noeviction`.",
			Optional:    true,
			Computed:    true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
				"noeviction",
				"allkeys-lru",
				"volatile-lru",
				"allkeys-random",
				"volatile-random",
				"volatile-ttl",
				"volatile-lfu",
				"allkeys-lfu",
			}, false)),
		},
		"redis_persistence": {
			Type:             schema.TypeString,
			Description:      "Redis persistence. When persistence is 'rdb', Redis does RDB dumps each 10 minutes if any key is changed. Also RDB dumps are done according to backup schedule for backup purposes. When persistence is 'off', no RDB dumps and backups are done, so data can be lost at any moment if service is restarted for any reason, or if service is powered off. Also service can't be forked.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"off", "rdb"}, false)),
		},
		"redis_timeout": {
			Type:        schema.TypeInt,
			Description: "Redis idle connection timeout in seconds. Default is 300.",
			Computed:    true,
			Optional:    true,
		},
		"redis_acl_channels_default": {
			Type:             schema.TypeString,
			Description:      "Default ACL for pub/sub channels used when Redis user is created. Determines default pub/sub channels' ACL for new users if ACL is not supplied. When this option is not defined, all_channels is assumed to keep backward compatibility. This option doesn't affect Redis configuration acl-pubsub-default.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"allchannels", "resetchannels"}, false)),
		},
	}
}
