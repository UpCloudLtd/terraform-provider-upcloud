package database

import (
	"context"
	"math"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   serviceDescription("MySQL"),
		CreateContext: resourceMySQLCreate,
		ReadContext:   resourceMySQLRead,
		UpdateContext: resourceMySQLUpdate,
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

func resourceMySQLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := d.Set("type", string(upcloud.ManagedDatabaseServiceTypeMySQL)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDatabaseCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	return resourceMySQLRead(ctx, d, meta)
}

func resourceMySQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatabaseRead(ctx, d, meta)
}

func resourceMySQLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := resourceDatabaseUpdate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	diags = append(diags, resourceMySQLRead(ctx, d, meta)...)
	return diags
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
					schemaRDBMSDatabaseCommonProperties(),
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
			Type:             schema.TypeInt,
			Description:      "The minimum amount of time in seconds to keep binlog entries before deletion. This may be extended for services that require binlog entries for longer than the default for example if using the MySQL Debezium Kafka connector.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(600, 86400)),
		},
		"connect_timeout": {
			Type:             schema.TypeInt,
			Description:      "The number of seconds that the mysqld server waits for a connect packet before responding with Bad handshake",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(2, 3600)),
		},
		"default_time_zone": {
			Type:             schema.TypeString,
			Description:      "Default server time zone as an offset from UTC (from -12:00 to +12:00), a time zone name, or `SYSTEM` to use the MySQL server default.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(2, 100)),
		},
		"group_concat_max_len": {
			Type:        schema.TypeInt,
			Description: "The maximum permitted result length in bytes for the `GROUP_CONCAT()` function.",
			Optional:    true,
			Computed:    true,
			// int max is lower than actual acceptable limit but as this field is integer
			// it's probably safer to use type limits here.
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(4, math.MaxInt)),
		},
		"information_schema_stats_expiry": {
			Type:             schema.TypeInt,
			Description:      "The time, in seconds, before cached statistics expire.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(900, 31536000)),
		},
		"innodb_ft_min_token_size": {
			Type:             schema.TypeInt,
			Description:      "Minimum length of words that are stored in an InnoDB `FULLTEXT` index.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 16)),
		},
		"innodb_ft_server_stopword_table": {
			Type:        schema.TypeString,
			Description: "This option is used to specify your own InnoDB `FULLTEXT` index stopword list for all InnoDB tables.",
			Optional:    true,
			Computed:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(regexp.MustCompile("^.+/.+$"), "must match '^.+/.+$' pattern e.g. 'db_name/table_name'")),
		},
		"innodb_lock_wait_timeout": {
			Type:             schema.TypeInt,
			Description:      "The length of time in seconds an InnoDB transaction waits for a row lock before giving up.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 3600)),
		},
		"innodb_log_buffer_size": {
			Type:             schema.TypeInt,
			Description:      "The size in bytes of the buffer that InnoDB uses to write to the log files on disk.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1048576, math.MaxInt)),
		},
		"innodb_online_alter_log_max_size": {
			Type:             schema.TypeInt,
			Description:      "The upper limit in bytes on the size of the temporary log files used during online DDL operations for InnoDB tables.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(65536, math.MaxInt)),
		},
		"innodb_print_all_deadlocks": {
			Type:        schema.TypeBool,
			Description: "When enabled, information about all deadlocks in InnoDB user transactions is recorded in the error log. Disabled by default.",
			Optional:    true,
			Computed:    true,
		},
		"innodb_rollback_on_timeout": {
			Type:        schema.TypeBool,
			Description: "When enabled a transaction timeout causes InnoDB to abort and roll back the entire transaction.",
			Optional:    true,
			Computed:    true,
		},
		"interactive_timeout": {
			Type:             schema.TypeInt,
			Description:      "The number of seconds the server waits for activity on an interactive connection before closing it.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(30, 604800)),
		},
		"internal_tmp_mem_storage_engine": {
			Type:             schema.TypeString,
			Description:      "The storage engine for in-memory internal temporary tables.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"TempTable", "MEMORY"}, false)),
		},
		"long_query_time": {
			Type:             schema.TypeInt,
			Description:      "The `slow_query_logs` work as SQL statements that take more than `long_query_time` seconds to execute. Default is `10s`",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 3600)),
		},
		"max_allowed_packet": {
			Type:             schema.TypeInt,
			Description:      "Size of the largest message in bytes that can be received by the server. Default is `67108864` (64M)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(102400, 1073741824)),
		},
		"max_heap_table_size": {
			Type:             schema.TypeInt,
			Description:      "Limits the size of internal in-memory tables. Also set `tmp_table_size`. Default is `16777216` (16M)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1048576, 1073741824)),
		},
		"net_read_timeout": {
			Type:             schema.TypeInt,
			Description:      "The number of seconds to wait for more data from a connection before aborting the read.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 3600)),
		},
		"net_write_timeout": {
			Type:             schema.TypeInt,
			Description:      "The number of seconds to wait for a block to be written to a connection before aborting the write.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 3600)),
		},
		"slow_query_log": {
			Type:        schema.TypeBool,
			Description: "Slow query log enables capturing of slow queries. Setting `slow_query_log` to false also truncates the `mysql.slow_log` table. Default is off",
			Optional:    true,
			Computed:    true,
		},
		"sort_buffer_size": {
			Type:             schema.TypeInt,
			Description:      "Sort buffer size in bytes for `ORDER BY` optimization. Default is `262144` (256K)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(32768, 1073741824)),
		},
		"sql_mode": {
			Type: schema.TypeString,
			Description: `Global SQL mode. Set to empty to use MySQL server defaults. 
			When creating a new service and not setting this field default SQL mode (strict, SQL standard compliant) will be assigned.`,
			Optional: true,
			Computed: true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(regexp.MustCompile("^[A-Z_]*(,[A-Z_]+)*$"), "must match '^[A-Z_]*(,[A-Z_]+)*$' pattern e.g. 'ANSI,TRADITIONAL'")),
		},
		"sql_require_primary_key": {
			Type: schema.TypeBool,
			Description: `Require primary key to be defined for new tables or old tables modified with ALTER TABLE and fail if missing. 
			It is recommended to always have primary keys because various functionality may break if any large table is missing them.`,
			Optional: true,
			Computed: true,
		},
		"tmp_table_size": {
			Type:             schema.TypeInt,
			Description:      "Limits the size of internal in-memory tables. Also set `max_heap_table_size`. Default is `16777216` (16M)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1048576, 1073741824)),
		},
		"version": {
			Type:             schema.TypeString,
			Description:      "MySQL major version",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"8"}, false)),
			ForceNew:         true,
		},
		"wait_timeout": {
			Type:             schema.TypeInt,
			Description:      "The number of seconds the server waits for activity on a noninteractive connection before closing it.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 2147483)),
		},
		"innodb_read_io_threads": {
			Type:             schema.TypeInt,
			Description:      "The number of I/O threads for read operations in InnoDB. Default is 4. Changing this parameter will lead to a restart of the MySQL service.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 64)),
		},
		"innodb_flush_neighbors": {
			Type:             schema.TypeInt,
			Description:      "Specifies whether flushing a page from the InnoDB buffer pool also flushes other dirty pages in the same extent (default is 1): 0 - dirty pages in the same extent are not flushed,  1 - flush contiguous dirty pages in the same extent,  2 - flush dirty pages in the same extent",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2)),
		},
		"innodb_change_buffer_max_size": {
			Description:      "Maximum size for the InnoDB change buffer, as a percentage of the total size of the buffer pool. Default is 25",
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 50)),
		},
		"net_buffer_length": {
			Description:      "Start sizes of connection buffer and result buffer. Default is 16384 (16K). Changing this parameter will lead to a restart of the MySQL service.",
			Type:             schema.TypeInt,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1024, 1048576)),
		},
		"innodb_thread_concurrency": {
			Type:             schema.TypeInt,
			Description:      "Defines the maximum number of threads permitted inside of InnoDB. Default is 0 (infinite concurrency - no limit)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 1000)),
		},
		"innodb_write_io_threads": {
			Type:             schema.TypeInt,
			Description:      "The number of I/O threads for write operations in InnoDB. Default is 4. Changing this parameter will lead to a restart of the MySQL service.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 64)),
		},
		"service_log": {
			Type:        schema.TypeBool,
			Description: "Store logs for the service so that they are available in the HTTP API and console.",
			Optional:    true,
			Computed:    true,
		},
	}
}
