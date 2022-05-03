package database

import (
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents MySQL managed database",
		CreateContext: resourceDatabaseCreate(managedDatabaseTypeMySQL),
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(),
			schemaMySQLEngine(),
		),
	}
}

func schemaMySQLEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"properties": {
			Description: "Database Engine properties for MySQL",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: utils.JoinSchemas(
					schemaDatabaseCommonProperties(),
					schemaMySQLProperties(),
				),
			},
		},
	}
}

func schemaMySQLProperties() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"binlog_retention_period": {
			Type:        schema.TypeInt,
			Description: "The minimum amount of time in seconds to keep binlog entries before deletion. This may be extended for services that require binlog entries for longer than the default for example if using the MySQL Debezium Kafka connector.",
			Optional:    true,
			Computed:    true,
		},
		"connect_timeout": {
			Type:        schema.TypeInt,
			Description: "connect_timeout",
			Optional:    true,
			Computed:    true,
		},
		"default_time_zone": {
			Type:        schema.TypeString,
			Description: "default_time_zone",
			Optional:    true,
			Computed:    true,
		},
		"group_concat_max_len": {
			Type:        schema.TypeInt,
			Description: "group_concat_max_len",
			Optional:    true,
			Computed:    true,
		},
		"information_schema_stats_expiry": {
			Type:        schema.TypeInt,
			Description: "information_schema_stats_expiry",
			Optional:    true,
			Computed:    true,
		},
		"innodb_ft_min_token_size": {
			Type:        schema.TypeInt,
			Description: "innodb_ft_min_token_size",
			Optional:    true,
			Computed:    true,
		},
		"innodb_ft_server_stopword_table": {
			Type:        schema.TypeString,
			Description: "innodb_ft_server_stopword_table",
			Optional:    true,
			Computed:    true,
		},
		"innodb_lock_wait_timeout": {
			Type:        schema.TypeInt,
			Description: "innodb_lock_wait_timeout",
			Optional:    true,
			Computed:    true,
		},
		"innodb_log_buffer_size": {
			Type:        schema.TypeInt,
			Description: "innodb_log_buffer_size",
			Optional:    true,
			Computed:    true,
		},
		"innodb_online_alter_log_max_size": {
			Type:        schema.TypeInt,
			Description: "innodb_online_alter_log_max_size",
			Optional:    true,
			Computed:    true,
		},
		"innodb_print_all_deadlocks": {
			Type:        schema.TypeBool,
			Description: "innodb_print_all_deadlocks",
			Optional:    true,
			Computed:    true,
		},
		"innodb_rollback_on_timeout": {
			Type:        schema.TypeBool,
			Description: "innodb_rollback_on_timeout",
			Optional:    true,
			Computed:    true,
		},
		"interactive_timeout": {
			Type:        schema.TypeInt,
			Description: "interactive_timeout",
			Optional:    true,
			Computed:    true,
		},
		"internal_tmp_mem_storage_engine": {
			Type:        schema.TypeString,
			Description: "internal_tmp_mem_storage_engine",
			Optional:    true,
			Computed:    true,
		},
		"long_query_time": {
			Type:        schema.TypeInt,
			Description: "long_query_time",
			Optional:    true,
			Computed:    true,
		},
		"max_allowed_packet": {
			Type:        schema.TypeInt,
			Description: "max_allowed_packet",
			Optional:    true,
			Computed:    true,
		},
		"max_heap_table_size": {
			Type:        schema.TypeInt,
			Description: "max_heap_table_size",
			Optional:    true,
			Computed:    true,
		},
		"net_read_timeout": {
			Type:        schema.TypeInt,
			Description: "net_read_timeout",
			Optional:    true,
			Computed:    true,
		},
		"net_write_timeout": {
			Type:        schema.TypeInt,
			Description: "net_write_timeout",
			Optional:    true,
			Computed:    true,
		},
		"slow_query_log": {
			Type:        schema.TypeBool,
			Description: "slow_query_log",
			Optional:    true,
			Computed:    true,
		},
		"sort_buffer_size": {
			Type:        schema.TypeInt,
			Description: "sort_buffer_size",
			Optional:    true,
			Computed:    true,
		},
		"sql_mode": {
			Type:        schema.TypeString,
			Description: "sql_mode",
			Optional:    true,
			Computed:    true,
		},
		"sql_require_primary_key": {
			Type:        schema.TypeBool,
			Description: "sql_require_primary_key",
			Optional:    true,
			Computed:    true,
		},
		"tmp_table_size": {
			Type:        schema.TypeInt,
			Description: "tmp_table_size",
			Optional:    true,
			Computed:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "MySQL major version",
			Optional:    true,
			Computed:    true,
		},
		"wait_timeout": {
			Type:        schema.TypeInt,
			Description: "wait_timeout",
			Optional:    true,
			Computed:    true,
		},
	}
}
