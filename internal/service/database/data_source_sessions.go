package database

import (
	"context"
	"time"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceSessionsMySQL() *schema.Resource {
	return &schema.Resource{
		Description: "Current sessions of a MySQL managed database",
		ReadContext: dataSourceSessionsMySQLRead,
		Schema: utils.JoinSchemas(
			schemaDataSourceSessionsCommon(),
			schemaDataSourceSessionsMySQL(),
		),
	}
}

func DataSourceSessionsPostgreSQL() *schema.Resource {
	return &schema.Resource{
		Description: "Current sessions of a PostgreSQL managed database",
		ReadContext: dataSourceSessionsPostgreSQLRead,
		Schema: utils.JoinSchemas(
			schemaDataSourceSessionsCommon(),
			schemaDataSourceSessionsPostgreSQL(),
		),
	}
}

func DataSourceSessionsValkey() *schema.Resource {
	return &schema.Resource{
		Description: "Current sessions of a Valkey managed database",
		ReadContext: dataSourceSessionsValkeyRead,
		Schema: utils.JoinSchemas(
			schemaDataSourceSessionsCommon(),
			schemaDataSourceSessionsValkey(),
		),
	}
}

func schemaDataSourceSessionsCommon() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"limit": {
			Description: "Number of entries to receive at most.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     10,
		},
		"offset": {
			Description: "Offset for retrieved results based on sort order.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
		},
		"order": {
			Description: "Order by session field and sort retrieved results. Limited variables can be used for ordering.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "query_duration:desc",
		},
		"service": {
			Description: "Service's UUID for which these sessions belongs to",
			Type:        schema.TypeString,
			Required:    true,
			Computed:    false,
		},
	}
}

func schemaDataSourceSessionsMySQL() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sessions": {
			Description: "Current sessions",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem:        schemaDatabaseSessionMySQL(),
		},
	}
}

func schemaDataSourceSessionsPostgreSQL() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sessions": {
			Description: "Current sessions",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem:        schemaDatabaseSessionPostgreSQL(),
		},
	}
}

func schemaDataSourceSessionsValkey() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sessions": {
			Description: "Current sessions",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem:        schemaDatabaseSessionValkey(),
		},
	}
}

func schemaDatabaseSessionMySQL() *schema.Resource {
	return &schema.Resource{
		Description: "MySQL session",
		Schema: map[string]*schema.Schema{
			"application_name": {
				Description: "Name of the application that is connected to this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"client_addr": {
				Description: "IP address of the client connected to this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datname": {
				Description: "Name of the database this service is connected to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"id": {
				Description: "Process ID of this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query": {
				Description: "Text of this service's most recent query. If state is active this field shows the currently executing query. In all other states, it shows an empty string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query_duration": {
				Description: "The active query current duration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "Current overall state of this service: active: The service is executing a query, idle: The service is waiting for a new client command.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"usename": {
				Description: "Name of the user logged into this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func schemaDatabaseSessionPostgreSQL() *schema.Resource {
	return &schema.Resource{
		Description: "PostgreSQL session",
		Schema: map[string]*schema.Schema{
			"application_name": {
				Description: "Name of the application that is connected to this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"backend_start": {
				Description: "Time when this process was started, i.e., when the client connected to the server.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"backend_type": {
				Description: "Type of current service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"backend_xid": {
				Description: "Top-level transaction identifier of this service, if any.",
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
			},
			"backend_xmin": {
				Description: "The current service's xmin horizon.",
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
			},
			"client_addr": {
				Description: "IP address of the client connected to this service. If this field is null, it indicates either that the client is connected via a Unix socket on the server machine or that this is an internal process such as autovacuum.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"client_hostname": {
				Description: "Host name of the connected client, as reported by a reverse DNS lookup of `client_addr`.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"client_port": {
				Description: "TCP port number that the client is using for communication with this service, or -1 if a Unix socket is used.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"datid": {
				Description: "OID of the database this service is connected to.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"datname": {
				Description: "Name of the database this service is connected to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"id": {
				Description: "Process ID of this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query": {
				Description: "Text of this service's most recent query. If state is active this field shows the currently executing query. In all other states, it shows the last query that was executed.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query_duration": {
				Description: "The active query current duration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query_start": {
				Description: "Time when the currently active query was started, or if state is not active, when the last query was started.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "Current overall state of this service: active: The service is executing a query, idle: The service is waiting for a new client command.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state_change": {
				Description: "Time when the state was last changed.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"usename": {
				Description: "Name of the user logged into this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"usesysid": {
				Description: "OID of the user logged into this service.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"wait_event": {
				Description: "Wait event name if service is currently waiting.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wait_event_type": {
				Description: "The type of event for which the service is waiting, if any; otherwise NULL.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"xact_start": {
				Description: "Time when this process' current transaction was started, or null if no transaction is active.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

func schemaDatabaseSessionValkey() *schema.Resource {
	return &schema.Resource{
		Description: "Valkey session",
		Schema: map[string]*schema.Schema{
			"active_channel_subscriptions": {
				Description: "Number of active channel subscriptions",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"active_database": {
				Description: "Current database ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"active_pattern_matching_channel_subscriptions": {
				Description: "Number of pattern matching subscriptions.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"application_name": {
				Description: "Name of the application that is connected to this service.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"client_addr": {
				Description: "Number of pattern matching subscriptions.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connection_age": {
				Description: "Total duration of the connection in nanoseconds.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"connection_idle": {
				Description: "Idle time of the connection in nanoseconds.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"flags": {
				Description: "A set containing flags' descriptions.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"flags_raw": {
				Description: "Client connection flags in raw string format.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"id": {
				Description: "Process ID of this session.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"multi_exec_commands": {
				Description: "Number of commands in a MULTI/EXEC context.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"output_buffer": {
				Description: "Output buffer length.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"output_buffer_memory": {
				Description: "Output buffer memory usage.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"output_list_length": {
				Description: "Output list length (replies are queued in this list when the buffer is full).",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"query": {
				Description: "The last executed command.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query_buffer": {
				Description: "Query buffer length (0 means no query pending).",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"query_buffer_free": {
				Description: "Free space of the query buffer (0 means the buffer is full).",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceSessionsMySQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return dataSourceSessionsRead(ctx, d, meta, upcloud.ManagedDatabaseServiceTypeMySQL)
}

func dataSourceSessionsPostgreSQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return dataSourceSessionsRead(ctx, d, meta, upcloud.ManagedDatabaseServiceTypePostgreSQL)
}

func dataSourceSessionsValkeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return dataSourceSessionsRead(ctx, d, meta, upcloud.ManagedDatabaseServiceTypeValkey)
}

func dataSourceSessionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}, serviceType upcloud.ManagedDatabaseServiceType) (diags diag.Diagnostics) {
	client := meta.(*service.Service)
	serviceID := d.Get("service").(string)

	limit := d.Get("limit").(int)
	offset := d.Get("offset").(int)
	order := d.Get("order").(string)

	db, err := client.GetManagedDatabase(ctx, &request.GetManagedDatabaseRequest{UUID: serviceID})
	if err != nil {
		return diag.FromErr(err)
	}

	if db.Type != serviceType {
		return diag.Errorf("Getting sessions for Managed Database %s failed: database type %s is not valid for this data source", serviceID, db.Type)
	}

	sessions, err := client.GetManagedDatabaseSessions(ctx, &request.GetManagedDatabaseSessionsRequest{
		UUID:   serviceID,
		Limit:  limit,
		Offset: offset,
		Order:  order,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceID)

	if err := d.Set("service", serviceID); err != nil {
		return diag.FromErr(err)
	}

	switch serviceType {
	case upcloud.ManagedDatabaseServiceTypeMySQL:
		if err := d.Set("sessions", buildSessionsMySQL(sessions.MySQL)); err != nil {
			return diag.FromErr(err)
		}
	case upcloud.ManagedDatabaseServiceTypePostgreSQL:
		if err := d.Set("sessions", buildSessionsPostgreSQL(sessions.PostgreSQL)); err != nil {
			return diag.FromErr(err)
		}
	case upcloud.ManagedDatabaseServiceTypeValkey:
		err := d.Set("sessions", buildSessionsValkey(sessions.Valkey))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func buildSessionsMySQL(sessions []upcloud.ManagedDatabaseSessionMySQL) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0)

	if len(sessions) == 0 {
		return maps
	}

	for _, session := range sessions {
		maps = append(maps, map[string]interface{}{
			"application_name": session.ApplicationName,
			"client_addr":      session.ClientAddr,
			"datname":          session.Datname,
			"id":               session.Id,
			"query":            session.Query,
			"query_duration":   session.QueryDuration.String(),
			"state":            session.State,
			"usename":          session.Usename,
		})
	}

	return maps
}

func buildSessionsPostgreSQL(sessions []upcloud.ManagedDatabaseSessionPostgreSQL) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0)

	if len(sessions) == 0 {
		return maps
	}

	for _, session := range sessions {
		maps = append(maps, map[string]interface{}{
			"application_name": session.ApplicationName,
			"backend_start":    session.BackendStart.UTC().Format(time.RFC3339Nano),
			"backend_type":     session.BackendType,
			"backend_xid":      session.BackendXid,
			"backend_xmin":     session.BackendXmin,
			"client_addr":      session.ClientAddr,
			"client_hostname":  session.ClientHostname,
			"client_port":      session.ClientPort,
			"datid":            session.Datid,
			"datname":          session.Datname,
			"id":               session.Id,
			"query":            session.Query,
			"query_duration":   session.QueryDuration.String(),
			"query_start":      session.QueryStart.UTC().Format(time.RFC3339Nano),
			"state":            session.State,
			"state_change":     session.StateChange.UTC().Format(time.RFC3339Nano),
			"usename":          session.Usename,
			"usesysid":         session.Usesysid,
			"wait_event":       session.WaitEvent,
			"wait_event_type":  session.WaitEventType,
			"xact_start":       session.XactStart.UTC().Format(time.RFC3339Nano),
		})
	}

	return maps
}

func buildSessionsValkey(sessions []upcloud.ManagedDatabaseSessionValkey) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0)

	if len(sessions) == 0 {
		return maps
	}

	for _, session := range sessions {
		maps = append(maps, map[string]interface{}{
			"active_channel_subscriptions":                  session.ActiveChannelSubscriptions,
			"active_database":                               session.ActiveDatabase,
			"active_pattern_matching_channel_subscriptions": session.ActivePatternMatchingChannelSubscriptions,
			"application_name":                              session.ApplicationName,
			"client_addr":                                   session.ClientAddr,
			"connection_age":                                session.ConnectionAge.Nanoseconds(),
			"connection_idle":                               session.ConnectionIdle.Nanoseconds(),
			"flags":                                         session.Flags,
			"flags_raw":                                     session.FlagsRaw,
			"id":                                            session.Id,
			"multi_exec_commands":                           session.MultiExecCommands,
			"output_buffer":                                 session.OutputBuffer,
			"output_buffer_memory":                          session.OutputBufferMemory,
			"output_list_length":                            session.OutputListLength,
			"query":                                         session.Query,
			"query_buffer":                                  session.QueryBuffer,
			"query_buffer_free":                             session.QueryBufferFree,
		})
	}

	return maps
}
