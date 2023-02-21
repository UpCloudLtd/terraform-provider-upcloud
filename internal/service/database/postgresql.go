package database

import (
	"math"
	"regexp"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePostgreSQL() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents PostgreSQL managed database",
		CreateContext: resourceDatabaseCreate(upcloud.ManagedDatabaseServiceTypePostgreSQL),
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: utils.JoinSchemas(
			schemaDatabaseCommon(),
			schemaPostgreSQLEngine(),
		),
	}
}

func schemaPostgreSQLEngine() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sslmode": {
			Description: "SSL Connection Mode for PostgreSQL",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"properties": {
			Description: "Database Engine properties for PostgreSQL",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: utils.JoinSchemas(
					schemaRDBMSDatabaseCommonProperties(),
					schemaDatabaseCommonProperties(),
					schemaPostgreSQLProperties(),
				),
			},
		},
	}
}

func schemaPostgreSQLProperties() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"autovacuum_analyze_scale_factor": {
			Type:             schema.TypeFloat,
			Description:      "Specifies a fraction of the table size to add to `autovacuum_analyze_threshold` when deciding whether to trigger an `ANALYZE`. The default is `0.2` (20% of table size)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.FloatBetween(0, 1)),
		},
		"autovacuum_analyze_threshold": {
			Type:             schema.TypeInt,
			Description:      "Specifies the minimum number of inserted, updated or deleted tuples needed to trigger an ANALYZE in any one table. The default is `50` tuples.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2147483647)),
		},
		"autovacuum_freeze_max_age": {
			Type: schema.TypeInt,
			Description: "Specifies the maximum age (in transactions) that a table's `pg_class.relfrozenxid` field can attain before a `VACUUM` operation " +
				`is forced to prevent transaction ID wraparound within the table. 
				Note that the system will launch autovacuum processes to prevent wraparound even when autovacuum is otherwise disabled. 
				This parameter will cause the server to be restarted.`,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(200000000, 1500000000)),
		},
		"autovacuum_max_workers": {
			Type: schema.TypeInt,
			Description: "Specifies the maximum number of autovacuum processes (other than the autovacuum launcher) that may be running at any one time. " +
				"The default is `3`. This parameter can only be set at server start.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 20)),
		},
		"autovacuum_naptime": {
			Type:             schema.TypeInt,
			Description:      "Specifies the minimum delay between autovacuum runs on any given database. The delay is measured in seconds, and the default is `1` minute",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 86400)),
		},
		"autovacuum_vacuum_cost_delay": {
			Type: schema.TypeInt,
			Description: "Specifies the cost delay value that will be used in automatic VACUUM operations. " +
				"If `-1` is specified, the regular `vacuum_cost_delay` value will be used. The default value is `20` milliseconds",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 100)),
		},
		"autovacuum_vacuum_cost_limit": {
			Type: schema.TypeInt,
			Description: "Specifies the cost limit value that will be used in automatic `VACUUM` operations. " +
				"If `-1` is specified (which is the default), the regular `vacuum_cost_limit` value will be used.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 10000)),
		},
		"autovacuum_vacuum_scale_factor": {
			Type: schema.TypeFloat,
			Description: "Specifies a fraction of the table size to add to autovacuum_vacuum_threshold when deciding whether to trigger a `VACUUM`. " +
				"The default is `0.2` (20% of table size)",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.FloatBetween(0, 1)),
		},
		"autovacuum_vacuum_threshold": {
			Type:             schema.TypeInt,
			Description:      "Specifies the minimum number of updated or deleted tuples needed to trigger a `VACUUM` in any one table. The default is `50` tuples",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2147483647)),
		},
		"bgwriter_delay": {
			Type:             schema.TypeInt,
			Description:      "Specifies the delay between activity rounds for the background writer in milliseconds. Default is `200`.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(10, 10000)),
		},
		"bgwriter_flush_after": {
			Type: schema.TypeInt,
			Description: "Whenever more than `bgwriter_flush_after` bytes have been written by the background writer, attempt to force the OS to issue these writes to the underlying storage. " +
				"Specified in kilobytes, default is `512`. Setting of `0` disables forced writeback.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2048)),
		},
		"bgwriter_lru_maxpages": {
			Type: schema.TypeInt,
			Description: "In each round, no more than this many buffers will be written by the background writer. " +
				"Setting this to zero disables background writing. Default is `100`.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 1073741823)),
		},
		"bgwriter_lru_multiplier": {
			Type: schema.TypeFloat,
			Description: "The average recent need for new buffers is multiplied by `bgwriter_lru_multiplier` to arrive at an estimate of the number that will be needed during the next round (up to `bgwriter_lru_maxpages`). " +
				"`1.0` represents a \"just in time\" policy of writing exactly the number of buffers predicted to be needed. " +
				"Larger values provide some cushion against spikes in demand, while smaller values intentionally leave writes to be done by server processes. " +
				"The default is `2.0`.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.FloatBetween(0, 10)),
		},
		"deadlock_timeout": {
			Type:             schema.TypeInt,
			Description:      "This is the amount of time, in milliseconds, to wait on a lock before checking to see if there is a deadlock condition.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(500, 1800000)),
		},
		"default_toast_compression": {
			Type:             schema.TypeString,
			Description:      "Controls the amount of detail written in the server log for each message that is logged.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"lz4", "pglz"}, false)),
		},
		"idle_in_transaction_session_timeout": {
			Type:             schema.TypeInt,
			Description:      "Time out sessions with open transactions after this number of milliseconds.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 604800000)),
		},
		"jit": {
			Type:        schema.TypeBool,
			Description: "Controls system-wide use of Just-in-Time Compilation (JIT).",
			Optional:    true,
			Computed:    true,
		},
		"log_autovacuum_min_duration": {
			Type: schema.TypeInt,
			Description: "Causes each action executed by autovacuum to be logged if it ran for at least the specified number of milliseconds. " +
				"Setting this to `0` logs all autovacuum actions. The default `-1` disables logging autovacuum actions.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 2147483647)),
		},
		"log_error_verbosity": {
			Type:             schema.TypeString,
			Description:      "Controls the amount of detail written in the server log for each message that is logged.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"TERSE", "DEFAULT", "VERBOSE"}, false)),
		},
		"log_line_prefix": {
			Type:        schema.TypeString,
			Description: "Choose from one of the available log-formats. These can support popular log analyzers like pgbadger, pganalyze etc.",
			Optional:    true,
			Computed:    true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
				"'%m [%p] %q[user=%u,db=%d,app=%a] '",
				"'%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '",
				"'pid=%p,user=%u,db=%d,app=%a,client=%h '",
			}, true)),
		},
		"log_min_duration_statement": {
			Type:             schema.TypeInt,
			Description:      "Log statements that take more than this number of milliseconds to run, `-1` disables",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 86400000)),
		},
		"max_files_per_process": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum number of files that can be open per process.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1000, 4096)),
		},
		"max_locks_per_transaction": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum locks per transaction.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(64, 6400)),
		},
		"max_logical_replication_workers": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum logical replication workers (taken from the pool of `max_parallel_workers`).",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(4, 64)),
		},
		"max_parallel_workers": {
			Type:             schema.TypeInt,
			Description:      "Sets the maximum number of workers that the system can support for parallel queries.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 96)),
		},
		"max_parallel_workers_per_gather": {
			Type:             schema.TypeInt,
			Description:      "Sets the maximum number of workers that can be started by a single Gather or Gather Merge node.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 96)),
		},
		"max_pred_locks_per_transaction": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum predicate locks per transaction.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(64, 5120)),
		},
		"max_prepared_transactions": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum prepared transactions",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 10000)),
		},
		"max_replication_slots": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum replication slots.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(8, 64)),
		},
		"max_slot_wal_keep_size": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum WAL size (MB) reserved for replication slots. Default is `-1` (unlimited). `wal_keep_size` minimum WAL size setting takes precedence over this.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 2147483647)),
		},
		"max_stack_depth": {
			Type:             schema.TypeInt,
			Description:      "Maximum depth of the stack in bytes.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(2097152, 6291456)),
		},
		"max_standby_archive_delay": {
			Type:             schema.TypeInt,
			Description:      "Max standby archive delay in milliseconds.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 43200000)),
		},
		"max_standby_streaming_delay": {
			Type:             schema.TypeInt,
			Description:      "Max standby streaming delay in milliseconds.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 43200000)),
		},
		"max_wal_senders": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL maximum WAL senders.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(20, 64)),
		},
		"max_worker_processes": {
			Type:             schema.TypeInt,
			Description:      "Sets the maximum number of background processes that the system can support.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(8, 96)),
		},
		"pg_partman_bgw_interval": {
			Type:             schema.TypeInt,
			Description:      "Sets the time interval to run pg_partman's scheduled tasks.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(3600, 604800)),
		},
		"pg_partman_bgw_role": {
			Type:        schema.TypeString,
			Description: "Controls which role to use for pg_partman's scheduled background tasks.",
			Optional:    true,
			Computed:    true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringMatch(regexp.MustCompile("^[_A-Za-z0-9][-._A-Za-z0-9]{0,63}$"), "must match '^[_A-Za-z0-9][-._A-Za-z0-9]{0,63}$' pattern e.g. 'myrolename'")),
		},
		"pg_read_replica": {
			Deprecated:  "Use read_replica service integration instead",
			Type:        schema.TypeBool,
			Description: "Should the service which is being forked be a read replica (deprecated, use read_replica service integration instead).",
			Optional:    true,
			Computed:    true,
		},
		"pg_service_to_fork_from": {
			Type:             schema.TypeString,
			Description:      "Name of the PG Service from which to fork (deprecated, use service_to_fork_from). This has effect only when a new service is being created.",
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressCreateOnlyProperty,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 64)),
		},
		"pg_stat_statements_track": {
			Type: schema.TypeString,
			Description: `Controls which statements are counted. 
			Specify top to track top-level statements (those issued directly by clients), all to also track nested statements (such as statements invoked within functions), 
			or none to disable statement statistics collection.` + "The default value is `top`.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"all", "top", "none"}, false)),
		},
		"shared_buffers_percentage": {
			Type: schema.TypeInt,
			Description: `Percentage of total RAM that the database server uses for shared memory buffers. 
				Valid range is 20-60 (float), which corresponds to 20% - 60%. ` + "This setting adjusts the `shared_buffers` configuration value.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(20, 60)),
		},
		"synchronous_replication": {
			Type:             schema.TypeString,
			Description:      "Synchronous replication type. Note that the service plan also needs to support synchronous replication.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"quorum", "off"}, false)),
		},
		"temp_file_limit": {
			Type:             schema.TypeInt,
			Description:      "PostgreSQL temporary file limit in KiB, -1 for unlimited",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 2147483647)),
		},
		"timezone": {
			Type:             schema.TypeString,
			Description:      "timezone",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 64)),
		},
		"track_activity_query_size": {
			Type:             schema.TypeInt,
			Description:      "Specifies the number of bytes reserved to track the currently executing command for each active session.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1024, 10240)),
		},
		"track_commit_timestamp": {
			Type:             schema.TypeString,
			Description:      "Record commit time of transactions.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"on", "off"}, false)),
		},
		"track_functions": {
			Type:             schema.TypeString,
			Description:      "Enables tracking of function call counts and time used.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"all", "pl", "none"}, false)),
		},
		"track_io_timing": {
			Type: schema.TypeString,
			Description: `Enables timing of database I/O calls. 
			This parameter is off by default, because it will repeatedly query the operating system for the current time, which may cause significant overhead on some platforms.`,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"on", "off"}, false)),
		},
		"variant": {
			Type:             schema.TypeString,
			Description:      "Variant of the PostgreSQL service, may affect the features that are exposed by default",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"aiven", "timescale"}, false)),
		},
		"version": {
			Type:             schema.TypeString,
			Description:      "PostgreSQL major version",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"12", "13", "14"}, false)),
		},
		"wal_sender_timeout": {
			Type:        schema.TypeInt,
			Description: "Terminate replication connections that are inactive for longer than this amount of time, in milliseconds. Setting this value to `0` disables the timeout.",
			Optional:    true,
			Computed:    true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.Any(
				validation.IntBetween(0, 0),
				validation.IntBetween(5000, 10800000)),
			),
		},
		"wal_writer_delay": {
			Type:             schema.TypeInt,
			Description:      "WAL flush interval in milliseconds. Note that setting this value to lower than the default `200`ms may negatively impact performance",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(10, 200)),
		},
		"work_mem": {
			Type: schema.TypeInt,
			Description: `Sets the maximum amount of memory to be used by a query operation (such as a sort or hash table) before writing to temporary disk files, 
			in MB. Default is 1MB + 0.075% of total RAM (up to 32MB).`,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 1024)),
		},
		"pgbouncer": {
			Description: "PGBouncer connection pooling settings",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem:        &schema.Resource{Schema: schemaPostgreSQLPGBouncer()},
		},
		"pglookout": {
			Description: "PGLookout settings",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_failover_replication_time_lag": {
						Type:             schema.TypeInt,
						Description:      "Number of seconds of master unavailability before triggering database failover to standby",
						Optional:         true,
						Default:          60,
						ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(10, math.MaxInt)),
					},
				},
			},
		},
		"timescaledb": {
			Description: "TimescaleDB extension configuration values",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_background_workers": {
						Type: schema.TypeInt,
						Description: `The number of background workers for timescaledb operations. 
						You should configure this setting to the sum of your number of databases and the total number of concurrent background workers you want running at any given point in time.`,
						Optional:         true,
						Computed:         true,
						ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 4096)),
					},
				},
			},
		},
		"pg_stat_monitor_pgsm_max_buckets": {
			Type:             schema.TypeInt,
			Description:      "Sets the maximum number of buckets ",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10)),
		},
		"pg_stat_monitor_pgsm_enable_query_plan": {
			Type:        schema.TypeBool,
			Description: "Enables or disables query plan monitoring",
			Optional:    true,
			Computed:    true,
		},
		"log_temp_files": {
			Type:             schema.TypeInt,
			Description:      "Log statements for each temporary file created larger than this number of kilobytes, -1 disables",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(-1, 2147483647)),
		},
		"pg_stat_monitor_enable": {
			Type:        schema.TypeBool,
			Description: "Enable the pg_stat_monitor extension. Enabling this extension will cause the cluster to be restarted.When this extension is enabled, pg_stat_statements results for utility commands are unreliable",
			Optional:    true,
			Computed:    true,
		},
	}
}

func schemaPostgreSQLPGBouncer() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"autodb_idle_timeout": {
			Type:             schema.TypeInt,
			Description:      "If the automatically created database pools have been unused this many seconds, they are freed. If 0 then timeout is disabled. [seconds]",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 86400)),
		},
		"autodb_max_db_connections": {
			Type:             schema.TypeInt,
			Description:      "Do not allow more than this many server connections per database (regardless of user). Setting it to 0 means unlimited.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 2147483647)),
		},
		"autodb_pool_mode": {
			Type:             schema.TypeString,
			Description:      "PGBouncer pool mode",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"session", "transaction", "statement"}, false)),
		},
		"autodb_pool_size": {
			Type:             schema.TypeInt,
			Description:      "If non-zero then create automatically a pool of that size per user when a pool doesn't exist.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 10000)),
		},
		"ignore_startup_parameters": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of parameters to ignore when given in startup packet",
			Optional:    true,
			Computed:    true,
		},
		"min_pool_size": {
			Type:             schema.TypeInt,
			Description:      "Add more server connections to pool if below this number. Improves behavior when usual load comes suddenly back after period of total inactivity. The value is effectively capped at the pool size.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 10000)),
		},
		"server_idle_timeout": {
			Type:             schema.TypeInt,
			Description:      "If a server connection has been idle more than this many seconds it will be dropped. If 0 then timeout is disabled. [seconds]",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 86400)),
		},
		"server_lifetime": {
			Type:             schema.TypeInt,
			Description:      "The pooler will close an unused server connection that has been connected longer than this. [seconds]",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(60, 86400)),
		},
		"server_reset_query_always": {
			Type:        schema.TypeBool,
			Description: "Run server_reset_query (`DISCARD ALL`) in all pooling modes",
			Optional:    true,
			Computed:    true,
		},
	}
}
