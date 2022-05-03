package database

import (
	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePostgreSQL() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource represents PostgreSQL managed database",
		CreateContext: resourceDatabaseCreate(managedDatabaseTypePostgreSQL),
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
			Type:        schema.TypeInt,
			Description: "autovacuum_analyze_scale_factor",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_analyze_threshold": {
			Type:        schema.TypeInt,
			Description: "autovacuum_analyze_threshold",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_freeze_max_age": {
			Type:        schema.TypeInt,
			Description: "autovacuum_freeze_max_age",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_max_workers": {
			Type:        schema.TypeInt,
			Description: "autovacuum_max_workers",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_naptime": {
			Type:        schema.TypeInt,
			Description: "autovacuum_naptime",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_vacuum_cost_delay": {
			Type:        schema.TypeInt,
			Description: "autovacuum_vacuum_cost_delay",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_vacuum_cost_limit": {
			Type:        schema.TypeInt,
			Description: "autovacuum_vacuum_cost_limit",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_vacuum_scale_factor": {
			Type:        schema.TypeInt,
			Description: "autovacuum_vacuum_scale_factor",
			Optional:    true,
			Computed:    true,
		},
		"autovacuum_vacuum_threshold": {
			Type:        schema.TypeInt,
			Description: "autovacuum_vacuum_threshold",
			Optional:    true,
			Computed:    true,
		},
		"bgwriter_delay": {
			Type:        schema.TypeInt,
			Description: "bgwriter_delay",
			Optional:    true,
			Computed:    true,
		},
		"bgwriter_flush_after": {
			Type:        schema.TypeInt,
			Description: "bgwriter_flush_after",
			Optional:    true,
			Computed:    true,
		},
		"bgwriter_lru_maxpages": {
			Type:        schema.TypeInt,
			Description: "bgwriter_lru_maxpages",
			Optional:    true,
			Computed:    true,
		},
		"bgwriter_lru_multiplier": {
			Type:        schema.TypeInt,
			Description: "bgwriter_lru_multiplier",
			Optional:    true,
			Computed:    true,
		},
		"deadlock_timeout": {
			Type:        schema.TypeInt,
			Description: "deadlock_timeout",
			Optional:    true,
			Computed:    true,
		},
		"idle_in_transaction_session_timeout": {
			Type:        schema.TypeInt,
			Description: "idle_in_transaction_session_timeout",
			Optional:    true,
			Computed:    true,
		},
		"jit": {
			Type:        schema.TypeBool,
			Description: "jit",
			Optional:    true,
			Computed:    true,
		},
		"log_autovacuum_min_duration": {
			Type:        schema.TypeInt,
			Description: "log_autovacuum_min_duration",
			Optional:    true,
			Computed:    true,
		},
		"log_error_verbosity": {
			Type:        schema.TypeString,
			Description: "log_error_verbosity",
			Optional:    true,
			Computed:    true,
		},
		"log_line_prefix": {
			Type:        schema.TypeString,
			Description: "log_line_prefix",
			Optional:    true,
			Computed:    true,
		},
		"log_min_duration_statement": {
			Type:        schema.TypeInt,
			Description: "log_min_duration_statement",
			Optional:    true,
			Computed:    true,
		},
		"max_files_per_process": {
			Type:        schema.TypeInt,
			Description: "max_files_per_process",
			Optional:    true,
			Computed:    true,
		},
		"max_locks_per_transaction": {
			Type:        schema.TypeInt,
			Description: "max_locks_per_transaction",
			Optional:    true,
			Computed:    true,
		},
		"max_logical_replication_workers": {
			Type:        schema.TypeInt,
			Description: "max_logical_replication_workers",
			Optional:    true,
			Computed:    true,
		},
		"max_parallel_workers": {
			Type:        schema.TypeInt,
			Description: "max_parallel_workers",
			Optional:    true,
			Computed:    true,
		},
		"max_parallel_workers_per_gather": {
			Type:        schema.TypeInt,
			Description: "max_parallel_workers_per_gather",
			Optional:    true,
			Computed:    true,
		},
		"max_pred_locks_per_transaction": {
			Type:        schema.TypeInt,
			Description: "max_pred_locks_per_transaction",
			Optional:    true,
			Computed:    true,
		},
		"max_prepared_transactions": {
			Type:        schema.TypeInt,
			Description: "max_prepared_transactions",
			Optional:    true,
			Computed:    true,
		},
		"max_replication_slots": {
			Type:        schema.TypeInt,
			Description: "max_replication_slots",
			Optional:    true,
			Computed:    true,
		},
		"max_stack_depth": {
			Type:        schema.TypeInt,
			Description: "max_stack_depth",
			Optional:    true,
			Computed:    true,
		},
		"max_standby_archive_delay": {
			Type:        schema.TypeInt,
			Description: "max_standby_archive_delay",
			Optional:    true,
			Computed:    true,
		},
		"max_standby_streaming_delay": {
			Type:        schema.TypeInt,
			Description: "max_standby_streaming_delay",
			Optional:    true,
			Computed:    true,
		},
		"max_wal_senders": {
			Type:        schema.TypeInt,
			Description: "max_wal_senders",
			Optional:    true,
			Computed:    true,
		},
		"max_worker_processes": {
			Type:        schema.TypeInt,
			Description: "max_worker_processes",
			Optional:    true,
			Computed:    true,
		},
		"pg_partman_bgw_interval": {
			Type:        schema.TypeInt,
			Description: "pg_partman_bgw.interval",
			Optional:    true,
			Computed:    true,
		},
		"pg_partman_bgw_role": {
			Type:        schema.TypeString,
			Description: "pg_partman_bgw.role",
			Optional:    true,
			Computed:    true,
		},
		"pg_read_replica": {
			Type:        schema.TypeBool,
			Description: "Should the service which is being forked be a read replica",
			Optional:    true,
			Computed:    true,
		},
		"pg_service_to_fork_from": {
			Type:             schema.TypeString,
			Description:      "Name of the PG Service from which to fork (deprecated, use service_to_fork_from). This has effect only when a new service is being created.",
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressCreateOnlyProperty,
		},
		"pg_stat_statements_track": {
			Type:        schema.TypeString,
			Description: "pg_stat_statements.track",
			Optional:    true,
			Computed:    true,
		},
		"shared_buffers_percentage": {
			Type:        schema.TypeInt,
			Description: "shared_buffers_percentage",
			Optional:    true,
			Computed:    true,
		},
		"synchronous_replication": {
			Type:        schema.TypeString,
			Description: "Synchronous replication type. Note that the service plan also needs to support synchronous replication.",
			Optional:    true,
			Computed:    true,
		},
		"temp_file_limit": {
			Type:        schema.TypeInt,
			Description: "temp_file_limit",
			Optional:    true,
			Computed:    true,
		},
		"timezone": {
			Type:        schema.TypeString,
			Description: "timezone",
			Optional:    true,
			Computed:    true,
		},
		"track_activity_query_size": {
			Type:        schema.TypeInt,
			Description: "track_activity_query_size",
			Optional:    true,
			Computed:    true,
		},
		"track_commit_timestamp": {
			Type:        schema.TypeString,
			Description: "track_commit_timestamp",
			Optional:    true,
			Computed:    true,
		},
		"track_functions": {
			Type:        schema.TypeString,
			Description: "track_functions",
			Optional:    true,
			Computed:    true,
		},
		"track_io_timing": {
			Type:        schema.TypeString,
			Description: "track_io_timing",
			Optional:    true,
			Computed:    true,
		},
		"variant": {
			Type:        schema.TypeString,
			Description: "Variant of the PostgreSQL service, may affect the features that are exposed by default",
			Optional:    true,
			Computed:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "PostgreSQL major version",
			Optional:    true,
			Computed:    true,
		},
		"wal_sender_timeout": {
			Type:        schema.TypeInt,
			Description: "wal_sender_timeout",
			Optional:    true,
			Computed:    true,
		},
		"wal_writer_delay": {
			Type:        schema.TypeInt,
			Description: "wal_writer_delay",
			Optional:    true,
			Computed:    true,
		},
		"work_mem": {
			Type:        schema.TypeInt,
			Description: "work_mem",
			Optional:    true,
			Computed:    true,
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
						Type:        schema.TypeInt,
						Description: "max_failover_replication_time_lag",
						Optional:    true,
						Default:     60,
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
						Type:        schema.TypeInt,
						Description: "timescaledb.max_background_workers",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

func schemaPostgreSQLPGBouncer() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"autodb_idle_timeout": {
			Type:        schema.TypeInt,
			Description: "If the automatically created database pools have been unused this many seconds, they are freed. If 0 then timeout is disabled. [seconds]",
			Optional:    true,
			Computed:    true,
		},
		"autodb_max_db_connections": {
			Type:        schema.TypeInt,
			Description: "Do not allow more than this many server connections per database (regardless of user). Setting it to 0 means unlimited.",
			Optional:    true,
			Computed:    true,
		},
		"autodb_pool_mode": {
			Type:        schema.TypeString,
			Description: "PGBouncer pool mode",
			Optional:    true,
			Computed:    true,
		},
		"autodb_pool_size": {
			Type:        schema.TypeInt,
			Description: "If non-zero then create automatically a pool of that size per user when a pool doesn't exist.",
			Optional:    true,
			Computed:    true,
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
			Type:        schema.TypeInt,
			Description: "Add more server connections to pool if below this number. Improves behavior when usual load comes suddenly back after period of total inactivity. The value is effectively capped at the pool size.",
			Optional:    true,
			Computed:    true,
		},
		"server_idle_timeout": {
			Type:        schema.TypeInt,
			Description: "If a server connection has been idle more than this many seconds it will be dropped. If 0 then timeout is disabled. [seconds]",
			Optional:    true,
			Computed:    true,
		},
		"server_lifetime": {
			Type:        schema.TypeInt,
			Description: "The pooler will close an unused server connection that has been connected longer than this. [seconds]",
			Optional:    true,
			Computed:    true,
		},
		"server_reset_query_always": {
			Type:        schema.TypeBool,
			Description: "Run server_reset_query (DISCARD ALL) in all pooling modes",
			Optional:    true,
			Computed:    true,
		},
	}
}
