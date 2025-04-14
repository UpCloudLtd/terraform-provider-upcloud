# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

## [5.21.0] - 2025-04-15

### Added

- upcloud_server: add `hot_resize` attribute, allowing plan changes without server restarts when supported by the platform. See [UpCloud documentation](https://upcloud.com/docs/guides/scale-cloud-servers-hot-resize/) for more information.

### Changed

- upcloud_managed_database_\*: Update available properties to match listing provided by the API, see #752 for details.

### Fixed

- upcloud_server: Correctly load configuration in Update method, fixing an issue where the `Host` field was not populated, potentially causing resource update failures.

## [5.20.5] - 2025-04-10

### Changed

- upcloud_managed_database_\*: Update available properties to match listing provided by the API, see #741 for details.

### Fixed

- Set `valkey_access_control` data when importing Valkey user to `upcloud_managed_database_user` resource.

## [5.20.4] - 2025-03-14

### Changed

- upcloud_server: set value for host field also when it is not defined by the user.

### Fixed

- Display descriptive error message when credentials are not configured.

## [5.20.3] - 2025-03-06

### Fixed

- upcloud_network: allow defining nexthop in `dhcp_routes`.

## [5.20.2] - 2025-03-04

### Fixed

- upcloud_kubernetes_cluster: remove client side default value for plan and use default value defined by API instead.
- upcloud_kubernetes_\*: remove the extra waits from delete methods. The back-end side has been updated so that cluster does not return HTTP 404 response until it has been fully removed.

## [5.20.1] - 2025-02-28

### Fixed

- Update labels validation: label key can contain printable ASCII characters and must not start with an underscore.

### Changed

- upcloud_managed_database_\*: Update available properties to match listing provided by the API, see #718 for details.

## [5.20.0] - 2025-02-24

### Added

- Experimental support for token authentication.

### Changed

- upcloud_managed_database_\*: Update available properties to match listing provided by the API, see #703 for details.

## [5.19.0] - 2025-02-13

### Added

- upcloud_loadbalancer_frontend_rule: add support for load balancer redirect rule status
- upcloud_hosts (Data Source): add `statistics` and `windows_enabled` fields

## [5.18.0] - 2025-01-30

### Changed

- upcloud_server: allow maximum of 31 additional_ip_address blocks instead of previous 4

## [5.17.0] - 2025-01-28

### Added

- upcloud_managed_database_postgresql: support for Postgres 17

### Changed

- upcloud_server: make template storage tier configurable.
- upcloud_managed_database_\*: Update available properties to match listing provided by the API, see #687 for details.

### Fixed

- upcloud_server: mark network interface IP address values as unknown during planning. This ensures that IP addresses have known values after apply.

## [5.16.0] - 2024-12-03

### Added

- upcloud_managed_database_\*: add `termination_protection` field.

## [5.15.0] - 2024-11-14

### Added

- upcloud_managed_object_storage_bucket resource for managing buckets in managed object storage services.
- upcloud_server: `index` field to `network_interfaces`.
- upcloud_managed_database_valkey: add support for Valkey.

### Changed

- upcloud_managed_database_\*: Update available properties to match listing provided by the API, see [#626](https://github.com/UpCloudLtd/terraform-provider-upcloud/pull/626) for details.
- upcloud_server: When modifying `network_interfaces`, match configured network interfaces to the server's actual network interfaces by `index` and `ip_address` (in addition to list order). This is to avoid public and utility network interfaces being re-assigned when the interfaces are re-ordered or when interface is removed from middle of the list. This might result to inaccurate diffs in the Terraform plan when interfaces are re-ordered or when interface is removed from middle of the list. We recommend explicitly setting the value for `index` in configuration, when interfaces are re-ordered or when interface is removed from middle of the list.

### Deprecated

- upcloud_managed_database_redis: Redis is deprecated in favor of Valkey. Please use Valkey for new key value store instances.

## [5.14.0] - 2024-10-28

### Changed

- Terraform: Introduce support for Terraform protocol version 6. Protocol version 6 requires Terraform CLI version 1.0 and later.

### Fixed

- upcloud_loadbalancer: Handling a changed value for `nodes` attribute on re-apply no longer causes an error.

## [5.13.2] - 2024-10-25

### Fixed

- upcloud_loadbalancer_frontend_rule: include `set_request_header`, and `set_response_header` in the _at least one action_ validator.

## [5.13.1] - 2024-10-24

### Fixed

- upcloud_managed_database_\*: Handle `["object", "null"]` property type (e.g. in `migration` property of MySQL databases) as an object when building API request to create and update DB properties.

## [5.13.0] - 2024-10-23

### Added

- Log UpCloud API requests and responses with debug level to Terraform logs.
- upcloud_loadbalancer_frontend_rule: add `http_status`, `http_status_range`, `request_header`, and `response_header` rule matchers.
- upcloud_loadbalancer_frontend_rule: add `set_request_header`, and `set_response_header` rule actions.

### Deprecated

- upcloud_loadbalancer_frontend_rule: `header` rule matcher. Use `request_header` rule matcher instead.

### Fixed

- upcloud_loadbalancer: allow `stopped` value to be set for `configured_status` field.

## [5.12.0] - 2024-10-10

### Added

- upcloud_loadbalancer_frontend_rule: added `matching_condition` field.
- provider: `ProviderConfigure()` and `NewWithUserAgent()` to allow setting user agent

### Fixed

- upcloud_router: use state for unknown `static_route.type` value as user defined routes always have `user` as type.

## [5.11.3] - 2024-10-09

### Changed

- upcloud_router: allow `no-nexthop` as static route nexthop value.

## [5.11.2] - 2024-10-02

### Fixed

- upcloud_loadbalancer_backend: added missing UpgradeState() to fix issue when upgrading the provider

## [5.11.1] - 2024-09-25

### Changed

- dependencies: bump `github.com/UpCloudLtd/upcloud-go-api` to `v8.8.1`

## [5.11.0] - 2024-09-11

### Added

- upcloud_storage (data source): `encrypt`, `labels` and `system_labels` read-only fields.
- upcloud_managed_object_storage_custom_domain resource for managing custom domains for managed object storage end-points.
- upcloud_load_balancer_dns_challenge_domain data source for configuring DNS settings required for validating certificates.

### Changed

- upcloud_storage (data source): allow using `id` and `title` fields to find the storage.
- upcloud_storage (data source): make `type` field optional.

### Deprecated

- upcloud_storage (data source): `name`, `name_regex` and `most_recent` fields.

### Fixed

- upcloud_storage: when uploading compressed image, set `Content-Type` header based on the file-extension of the path defined in `source_location`.

### Removed

- upcloud_loadbalancer_backend: `tls_configs` removed from properties. The computed field exists on main level of the resource.

## [5.10.1] - 2024-08-21

### Fixed

- upcloud_kubernetes_node_group: do not set default value for `storage_encryption` in the provider implementation. Instead, use value returned from the API when updating the node group state. This fixes data consistency error when creating node group without defining value for `storage_encryption` and does not try to replace node group when running apply after updating provider to `v5.10.0`.
- upcloud_network_peering: add errors to diagnostics correctly, if fetching peering details fails while waiting for peering state, instead of crashing due to a segfault.

## [5.10.0] - 2024-08-19

### Added

- upcloud_storage: add support for labels
- upcloud_storage_template: add support for creating custom templates
- upcloud_kubernetes_node_group: `standard` storage tier when using a custom plan
- upcloud_managed_database_opensearch: `azure_migration`, `gcs_migration`, `index_rollup`, and `s3_migration` properties.

### Changed

- upcloud_managed_object_storage_policy: store configured value instead of the value returned by the API in the Terraform state. The provider will raise an error if these documents do not match
- System defined labels (i.e. labels prefixed with `_`) are filtered out from the `labels` maps.

### Fixed

- upcloud_storage: use `source_hash` to automatically verify the integrity of the imported data. Previously, the value was stored to state, but no validations were done
- upcloud_managed_object_storage_policy: ignore whitespace and unnecessary escapes when determining if policy document has changed

## [5.9.1] - 2024-08-05

### Added

- upcloud_managed_database_mysql: `ignore_roles` property (supported by PostgreSQL only at the moment)
- upcloud_managed_database_postgresql: `ignore_roles` property
- upcloud_managed_database_postgresql: `max_prepared_statements` property
- upcloud_managed_database_redis: `ignore_roles` property (supported by PostgreSQL only at the moment)
- upcloud_storage: add `standard` as a supported storage tier

### Fixed

- upcloud_managed_object_storage: modifying `region` requires replacing the resource.

## [5.9.0] - 2024-07-29

### Added

- upcloud_managed_database_\*: support for labels
- upcloud_router: add `static_routes` set for listing both user and service defined static routes

### Changed

- upcloud_router: store `attached_networks` values in alphabetical order
- upcloud_router: do not include service defined routes in the `static_route` set, as those can not be modified or removed by the user

### Fixed

- upcloud_router: remove empty strings from `attached_networks` value

## [5.8.1] - 2024-07-18

### Fixed

- upcloud_storage: sync title length constraint with API, allows 1-255 characters now

## [5.8.0] - 2024-07-16

### Added

- upcloud_kubernetes_node_group: support for non-encrypted node groups in encrypted cluster
- upcloud_managed_database_opensearch: `knn_memory_circuit_breaker_enabled` and `knn_memory_circuit_breaker_limit` properties.

### Changed

- upcloud_loadbalancer_frontend: use set type for `networks` as the backend returns them in alphabetical order instead of maintaining the order
- upcloud_loadbalancer_frontend: only store networks in state when the networks have been configured using `networks` blocks instead of deprecated `upcloud_loadbalancer.network` field.

### Fixed

- upcloud_loadbalancer_frontend: handle changes in the `networks`
- upcloud_loadbalancer: set `maintenance_dow` and `maintenance_time` as computed to avoid planning them to be removed when missing from configuration.

## [5.7.0] - 2024-07-02

### Added

- upcloud_managed_database_postgresql: support for Postgres 16

### Fixed

- upcloud_kubernetes_cluster (data source): make `kubeconfig` value sensitive.

### Removed

- upcloud_managed_database_postgresql: `pgaudit` property

## [5.6.1] - 2024-06-25

### Fixed

- dependencies: bump `github.com/hashicorp/go-retryablehttp` to `v0.7.7` to avoid potentially leaking basic auth credentials to logs.

## [5.6.0] - 2024-06-19

### Added

- upcloud_router: support `labels` field

### Fixed

- upcloud_network: detect if resource was deleted outside of Terraform
- upcloud_network_peering: detect if resource was deleted outside of Terraform
- upcloud_floating_ip_address: replace floating IP address, if `family` or `zone` have changes.
- provider: do not replace zero value with default when configuring plugin framework provider. For example, `request_timeout_sec = 0` will now disable request timeout also for resources migrated to plugin framework (e.g. `upcloud_network`).

## [5.5.0] - 2024-06-04

### Added

- upcloud_zone: `parent_zone` field.
- upcloud_network: support `labels` field
- gateway: uuid field for `upcloud_gateway_connection` resource
- gateway: uuid field for `upcloud_gateway_connection_tunnel` resource

### Fixed

- upcloud_managed_database_\*: do not include empty or unmodified nested properties values in the API request. I.e., filter nested properties similarly than main level properties.

### Deprecated

- upcloud_zone: `name` field will be removed as the same value is available through `id` field.

## [5.4.0] - 2024-05-21

### Added

- kubernetes: support for node group custom plans
- kubernetes: support for node storage data at rest encryption

## [5.3.0] - 2024-05-13

### Added

- upcloud_network_peering: support for network peerings.

### Removed

- upcloud_managed_database_opensearch: `max_index_count` property in favor of `index_patterns`

### Fixed

- upcloud_managed_object_storage_user_policy: fix error with handling when policy is not found

## [5.2.3] - 2024-04-30

### Fixed

- upcloud_managed_object_storage_user_policy: fix issue with refreshing state after the resouce was deleted outside of Terraform
- upcloud_firewall_rules: fix issue with refreshing state after the resouce was deleted outside of Terraform
- upcloud_managed_database: removed additional properties are not stored in the state even if API returns them
- upcloud_managed_database_logical_database: fix issue with refreshing state after the resource was deleted outside of Terraform

## [5.2.2] - 2024-04-15

### Added

- upcloud_server: `additional_ip_address` block underneath `network_interface` for adding a maximum of 4 IP addresses to a server network interface

### Fixed

- managed object storage: fix error when refreshing a state after the resource was deleted outside of Terraform; applies to all upcloud_managed_object_storage\* resources

### Changed

- Go version bump to 1.21

## [5.2.1] - 2024-03-28

### Fixed

- upcloud_managed_database: do not populate removed properties after upgrading the provider

## [5.2.0] - 2024-03-26

### Added

- upcloud_gateway: support for VPN feature (note that VPN feature is currently in beta)
- upcloud_gateway: `upcloud_gateway_connection` resource for creating VPN connections
- upcloud_gateway: `upcloud_gateway_connection_tunnel` resource for creating VPN tunnels

## [5.1.1] - 2024-03-13

### Changed

- upcloud_managed_database: update properties for each database type to match upstream.

### Fixed

- upcloud_managed_database: set all fields when importing database resources
- docs: update provider version to `~> 5.0`

## [5.1.0] - 2024-03-07

### Added

- upcloud_managed_database: support for attaching private networks

### Fixed

- upcloud_managed_database: set all relevant fields when importing `logical_database` and `user` resources

## [5.0.3] - 2024-03-05

### Fixed

- upcloud_managed_object_storage: fix import of `user*` and `policy` resources

## [5.0.2] - 2024-03-04

### Fixed

- upcloud_managed_object_storage: set `service_uuid` on import based on the given id

## [5.0.1] - 2024-03-01

### Fixed

- Added missing data sources and resources to Terraform provider documentation

## [5.0.0] - 2024-02-29

### Added

- upcloud_managed_object_storage: `upcloud_managed_storage_policy` resource for setting up policies
- upcloud_managed_object_storage: `upcloud_managed_storage_policies` data source for policies
- upcloud_managed_object_storage: `upcloud_managed_storage_user` resource for user management
- upcloud_managed_object_storage: `upcloud_managed_storage_user_policy` resource for attaching policies to users
- upcloud_managed_object_storage: `iam_url` property to `upcloud_managed_storage.endpoint`
- upcloud_managed_object_storage: `sts_url` property to `upcloud_managed_storage.endpoint`
- upcloud_managed_object_storage: required `status` property to `upcloud_managed_storage_user_access_key`

### Removed

- **Breaking**, upcloud_managed_object_storage: `users` property from `upcloud_managed_storage` resource
- **Breaking**, upcloud_managed_object_storage: `enabled` property from `upcloud_managed_storage_user_access_key` resource
- **Breaking**, upcloud_managed_object_storage: `name` property from `upcloud_managed_storage_user_access_key` resource
- **Breaking**, upcloud_managed_object_storage: `updated_at` property from `upcloud_managed_storage_user_access_key` resource

## [4.1.0] - 2024-02-16

### Added

- upcloud_loadbalancer: `maintenance_dow` and `maintenance_time` for managing maintenance windows settings
- upcloud_kubernetes_cluster: support for labels

### Fixed

- docs: update provider version to `~> 4.0`

## [4.0.0] - 2024-01-26

### Added
- upcloud_managed_database_redis: `redis_version` property

### Changed

- **Breaking**, upcloud_managed_database_mysql: changing property `admin_password` or `admin_username` forces resource re-creation
- **Breaking**, upcloud_managed_database_postgresql: changing property `admin_password` or `admin_username` forces resource re-creation
- **Breaking**, upcloud_managed_database resources: `title` field is required

### Removed

- **Breaking**, upcloud_managed_database_postgresql: `pg_read_replica` property
- **Breaking**, upcloud_managed_database_postgresql: `pg_service_to_fork_from` property

## [3.4.0] - 2024-01-25

### Added

- server: Add `server_group` field to allow configuring anti-affinity group when creating the server.
- upcloud_loadbalancer_frontend_rule: add `inverse` option to rule matchers.
- storage: Add support for encryption.

## [3.3.1] - 2024-01-10

### Added

- docs: add links to related UpCloud product documentation and tutorials.

### Fixed

- docs: update provider version to `~> 3.0`
- docs: in `upcloud_kubernetes_node_group` example, fix references and add missing required parameters

## [3.3.0] - 2023-12-20

### Added

- managed_object_storage: add required `name` property

### Fixed

- managed_object_storage: support for not configuring `labels`

## [3.2.0] - 2023-12-19

### Added

- load_balancer: `upcloud_loadbalancer_backend_tls_config` resource for backend TLS config management
- load_balancer: fields `tls_enabled`, `tls_verify` & `tls_use_system_ca` to `upcloud_loadbalancer_backend` resource's `properties`
- load_balancer: `http2_enabled` to `upcloud_loadbalancer_backend` resource's `properties` for enabling HTTP/2 backend support
- managed_database_mysql: Add `service_log` property
- managed_database_postgresql: Add `service_log` property
- managed_database_redis: Add `service_log` property
- server: Add `address_position` field to `storage_devices` and `template`
- provider: `request_timeout_sec` field to `upcloud` provider for managing the duration (in seconds) that the provider waits for an HTTP request towards UpCloud API to complete. Defaults to 120 seconds.

## [3.1.1] - 2023-11-21

### Changed
- docs: group resources and data-sources by product

## [3.1.0] - 2023-11-09

### Added
- kubernetes: `version` field to `upcloud_kubernetes_cluster` resource

### Deprecated
- `upcloud_object_storage`: the target product will reach its end of life by the end of 2024.

## [3.0.3] - 2023-10-31

### Fixed
- kubernetes: `upcloud_kubernetes_node_group` resource re-creation waits for destruction before creation
- managed_object_storage: `upcloud_managed_object_storage` resource network related documentation improved
- ip: `upcloud_floating_ip_address` resource's `access` field to allow only `public` value

## [3.0.2] - 2023-10-24

### Fixed
- managed_object_storage: `upcloud_managed_object_storage` resource public network validation

## [3.0.1] - 2023-10-23

### Fixed
- managed_object_storage: `upcloud_managed_object_storage` resource update to retain service users in all cases

## [3.0.0] - 2023-10-23

### Added
- **Breaking**, kubernetes: `control_plane_ip_filter` field to `upcloud_kubernetes_cluster` resource. This changes default behavior from _allow access from any IP_ to _block access from all IPs_. To be able to connect to the cluster, define list of allowed IP addresses and/or CIDR blocks or allow access from any IP.
- gateway: add read-only `addresses` field
- dbaas: `upcloud_managed_database_mysql_sessions`, `upcloud_managed_database_postgresql_sessions` and `upcloud_managed_database_redis_sessions` data sources
- network: `dhcp_routes` field to `ip_network` block in `upcloud_network` resource
- router: `static_routes` block to `upcloud_router` resource
- managed_object_storage: `managed_object_storage` & `managed_object_storage_user_access_key` resources and `managed_object_storage_regions` data source

### Changed
- kubernetes: remove node group maximum value validation. The maximum number of nodes (in the cluster) is determined by the cluster plan and the validation is done on the API side.

### Fixed
- **Breaking**, server: change tags from `List` to `Set`. The list order has already been ignored earlier and API does not support defining the order of tags.
- servergroup: use valid value as default for `anti_affinity_policy`.

## [2.12.0] - 2023-07-21

### Added
- lbaas: add `health_check_tls_verify` field to backend properties
- kubernetes: `utility_network_access` field to `upcloud_kubernetes_node_group` resource

## [2.11.0] - 2023-06-07

### Added
- kubernetes: `private_node_groups` field to `upcloud_kubernetes_cluster` resource
- server: properties `timezone`, `video_model` and `nic_model`
- servergroup: `anti_affinity_policy` field to `upcloud_server_group` resource for supporting strict anti-affinity
- dbaas: `upcloud_managed_database_opensearch` resource
- dbaas: `opensearch_access_control` block to `upcloud_managed_database_user` resource
- dbaas: `upcloud_managed_database_opensearch_indices` data source

### Changed
- dbaas: modifying `upcloud_managed_database_mysql` resource version field forces a new resource

### Removed
- servergroup: `anti_affinity` field from `upcloud_server_group` in favor of anti_affinity_policy

## [2.10.0] - 2023-04-26

### Added
- kubernetes: plan field to `upcloud_kubernetes_cluster` resource
- dbaas: support for PostgreSQL version 15

### Changed
- update upcloud-go-api to v6.1.0

## [2.9.1] - 2023-04-03

### Added
- lbaas: add `labels` support
- server, server group: add validation for `labels` keys and values

### Fixed
- gateway: wait for gateway to reach running state during resource create

## [2.9.0] - 2023-03-13

### Added
- gateway: new `upcloud_gateway` resource

### Changed
- update upcloud-go-api to v6.0.0

## [2.8.4] - 2023-02-21

### Fixed
- kubernetes: `upcloud_kubernetes_cluster` data source now provides `client_certificate`, `client_key`, and `cluster_ca_certificate` as PEM strings instead of base64 encoded PEM strings

## [2.8.3] - 2023-01-31

### Added
- kubernetes: anti-affinity option for `upcloud_kubernetes_node_group` resource

### Changed
- update upcloud-go-api to v5.4.0

## [2.8.2] - 2023-01-30

### Added
- server: support for `daily` simple backup plan

## [2.8.1] - 2023-01-26

### Added
- kubernetes: experimental `upcloud_kubernetes_node_group` resource

### Changed
- update upcloud-go-api to v5.2.1

### Removed
- kubernetes: experimental `node_group` field from `upcloud_kubernetes_cluster` resource
- dbaas: properties `additional_backup_regions`, `enable_ipv6` and `recovery_basebackup_name`

## [2.8.0] - 2022-12-21

### Added
- dbaas: experimental support for Managed Redis Database
- dbaas: user ACL properties for Redis and PostgreSQL
- dbaas: MySQL user authentication type field
- lbaas: `scheme` field to frontend rule HTTP redirect action.

### Changed
- update upcloud-go-api to v5.2.0

## [2.7.1] - 2022-11-29

### Fixed
- dbaas: new DB properties causing error when updating the state

### Added
- dbaas: new properties to MySQL and PostgreSQL resources

### Changed
- server: rebuild network interfaces without re-creating server
- new upcloud-go-api version 5

## [2.7.0] - 2022-11-16

### Added
- lbaas: private network support
- new server group resource with experimental anti affinity support

### Changed
- Update terraform-plugin-sdk to v2.24.0
- Update upcloud-go-api to v4.10.0

### Deprecated
- lbaas: `upcloud_loadbalancer` resource fields `dns_name` and `network`

### Removed
- kubernetes: kubernetes plan datasource

## [2.6.1] - 2022-10-12

### Fixed
- kubernetes: add mention about k8s resources being in alpha to the docs

## [2.6.0] - 2022-10-11

### Added
- dbaas: property validators
- dbaas: PostgreSQL properties `default_toast_compression` and `max_slot_wal_keep_size`
- server: Labels
- kubernetes: experimental: `upcloud_kubernetes_cluster` resource and `upcloud_kubernetes_cluster`
  & `upcloud_kubernetes_plan` data sources

### Fixed
- dbaas: fractional values in PostgreSQL properties `autovacuum_analyze_scale_factor`, `autovacuum_vacuum_scale_factor` and `bgwriter_lru_multiplier`
- dbaas: removed logic to replace last `_` with `.` in `pg_stat_statements_track`, `pg_partman_bgw_role`, `pg_partman_bgw_interval` property names as these are now handled in snake case also in the API.

### Changed
- dbaas: updated property descriptions
- structured logging with `tflog`
- storage: update maximum storage size from 2048 to 4096 gigabytes
- provider: changed `username` and `password` into optional parameters. This does not change how these parameters are used: providing these values in the provider block has already been optional, if credentials were defined as environment variables.

## [2.5.0] - 2022-06-20

### Added
- lbaas: frontend and backend properties
- lbaas: `set_forwarded_headers` frontend rule action
- firewall: allow specifying default rules

### Changed
- New upcloud-go-api version 4.7.0 with context support

## [2.4.2] - 2022-05-10

### Changed
- Update GoReleaser to v1.8.3

## [2.4.1] - 2022-05-05

### Fixed

- server: Remove all tags when tags change into an empty value
- server: Delete unused tags created with server resource on server delete
- server: Improve tags validation: check for case-insensitive duplicates, supress diff when only order of tags changes, print warning when trying to create existing tag with different letter casing
- dbaas: require that both `maintenance_window_time` and `maintenance_window_dow` are set when defining maintenance window
- dbaas: `maintenance_window_time` format

### Changed
- New upcloud-go-api version v4.5.1
- Update terraform-plugin-sdk to v2.15.0
- Update Go version to 1.17

## [2.4.0] - 2022-04-12

### Added
- Support for UpCloud Managed Load Balancers (beta)

### Fixed
- dbaas: upgrading database version

## [2.3.0] - 2022-03-14

### Added
- object storage: allow passing access and secret key as environment variables
- object storage: enable import feature
- storage: add support for autoresizing partition and filesystem

### Fixed
- dbaas: fix PostgreSQL properties: pg_stat_statements_track, pg_partman_bgw_role, pg_partman_bgw_interval

## [2.2.0] - 2022-02-14

### Added

- storage: upcloud_storage data source to retrieve specific storage details

### Fixed

- docs: set provider username and password as required arguments
- provider: return underlying error from initial login check instead of custom error
- provider: fix dangling resource references by removing a binding to an remote object if it no longer exists
- provider: fix runtime error when importing managed database


## [2.1.5] - 2022-01-27

### Fixed

- storage: fix missing backup_rule when importing resource
- provider: fix user-agent for release builds
- server: fix missing template id if resource creation fails on tag errors

### Changed

- Update documentation


## [2.1.4] - 2022-01-18

### Added

- server: validate plan and zone field values before executing API commands
- Support for UpCloud Managed Databases
- Support for debuggers like Delve

### Fixed

- firewall: fix missing server_id when importing firewall resource
- firewall: change port types from int to string to avoid having zero values in state when importing rules with undefined port number(s).
- firewall: remove proto field's default value "tcp" as this prevents settings optional fields value to null and update validator to accept empty string which corresponds to any protocol
- object storage: fix issue where order of storage buckets in an object storage resource would incorrectly trigger changes
- server: return more descriptive error message if subaccount tries to edit server tags

### Changed

- Upgraded terraform-plugin-sdk from v2.7.1 to v2.10.0

## [2.1.3] - 2021-11-18

### Added

- Added title field to the server resource

### Fixed

- server: fix custom plan updates (cpu/mem)

### Changed

- server: new hostname validator

## [2.1.2] - 2021-11-01

### Added

- Added simple backups support (#188)

### Fixed

- Prevent empty tags from replanning a server (#178)
- Make sure either storage devices or template are required on the server resource

## [2.1.1] - 2021-06-22

### Fixed

- fix(client): fix user-agent value (#165)

## [2.1.0] - 2021-06-01

### Added

- Support for UpCloud ObjectStorage S3 compatible storage.
- Add host field to the server resource
- server: add tags attribute support (#150)
- chore: Add more examples

### Fixed

- Server not started after updating storage device
- router: fix creation of attachedNetworks for routers #144
- chore: fix example in upcloud_tag #125
- server: prevent some attribute update from restarting (#146)
- router: allow detaching router and deleting attached routers (#151)
- storage: check size before cloning a device (#152)
- storage: fix address formating (#153)

### Changed

- Update documentation
- Update README

### Deprecated

- tag resource
- zone and zones datasources
- tag datasource

## [2.0.0] - 2021-01-27

### Added

- Missing documentation server resource [#89](https://github.com/UpCloudLtd/terraform-provider-upcloud/issues/89)
- Missing documentation for zone datasource [#120](https://github.com/UpCloudLtd/terraform-provider-upcloud/issues/120)
- New [examples](https://github.com/UpCloudLtd/terraform-provider-upcloud/tree/main/examples/) of using the provider
- Updated workflow to run acceptance tests when opening pull request / pushing to master
- Add user-agent header to the requests
- Can now explicitly set IP address for network interfaces (requires special priviledes for your UpCloud account)
- Expose metadata field for server resource

### Changed

- **Breaking**: the template (os storage) is described with a separate block within the server resource, note that removing / recreating server resource also recreates the storage
- **Breaking**: other storages are now managed outside of the server resource and attached to server using `storage_devices` block

### Removed

- Moved multiple utility functions to `/internal`

### Fixed

- Better drift detection [#106](https://github.com/UpCloudLtd/terraform-provider-upcloud/issues/106)
- Fixed issue where a change in server storages would replace the server network interfaces and recreate the server
- Addressed issue where a change in server networking would replace the previous storages (the template will still be created anew)
- Inconsistent documentation

## [1.0.0] - 2020-10-19

Updated upcloud-go-api, added build/CI scripts, and repackaged 0.1.0 as 1.0.0.

## [0.1.0] - 2020-09-24

### Added

- Changelog to highlight key alterations across future releases
- Website directory for future provider documentation
- Vendor directory through go modules to cover CI builds
- datasource_upcloud_hosts to view hosts data
- datasource_upcloud_ip_addresses to retrieve account wide ip address data
- datasource_upcloud_networks to retrieve account wide networks data
- datasource_upcloud_tags to retrieve account wide tag data
- datasource_upcloud_zone to retrieve specific zone details
- datasource_upcloud_zones to retrieve account wide zone data
- resource_upcloud_firewall_rules add to allow rules to be applied to server
- resource_upcloud_floating_ip_address to allow the management of floating ip addresses
- resource_upcloud_network to allow the management of networks
- resource_upcloud_router to allow the management of routers

### Changed

- README and examples/README to cover local builds, setup and test execution
- Go version to 1.14 and against Go master branch in Travis CI
- Travis CI file to execute website-test covering provider documentation
- Provider uses Terraform Plugin SDK V2
- resource_upcloud_server expanded with new functionality from UpCloud API 1.3
- resource_upcloud_storage expaned with new functionality from UpCloud API 1.3
- resource_upcloud_tag expanded to implement read function

### Removed

- Removed storage HCL blocks that failed due to referencing older UpCloud template ID
- Removed the plan, price, price_zone and timezone UpCloud resources
- resource_upcloud_ip removed and replaced by resource_upcloud_floating_ip_address
- resource_upcloud_firewall_rule removed and replaced by resource_upcloud_firewall_rules
- resource_upcloud_zone removed and replaced by zone and zones datasources

[Unreleased]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.21.0...HEAD
[5.21.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.20.5...v5.21.0
[5.20.5]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.20.4...v5.20.5
[5.20.4]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.20.3...v5.20.4
[5.20.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.20.2...v5.20.3
[5.20.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.20.1...v5.20.2
[5.20.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.20.0...v5.20.1
[5.20.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.19.0...v5.20.0
[5.19.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.18.0...v5.19.0
[5.18.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.17.0...v5.18.0
[5.17.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.16.0...v5.17.0
[5.16.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.15.0...v5.16.0
[5.15.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.14.0...v5.15.0
[5.14.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.13.2...v5.14.0
[5.13.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.13.1...v5.13.2
[5.13.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.13.0...v5.13.1
[5.13.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.12.0...v5.13.0
[5.12.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.11.3...v5.12.0
[5.11.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.11.2...v5.11.3
[5.11.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.11.1...v5.11.2
[5.11.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.11.0...v5.11.1
[5.11.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.10.1...v5.11.0
[5.10.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.10.0...v5.10.1
[5.10.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.9.1...v5.10.0
[5.9.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.9.0...v5.9.1
[5.9.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.8.1...v5.9.0
[5.8.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.8.0...v5.8.1
[5.8.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.7.0...v5.8.0
[5.7.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.6.1...v5.7.0
[5.6.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.6.0...v5.6.1
[5.6.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.5.0...v5.6.0
[5.5.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.4.0...v5.5.0
[5.4.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.3.0...v5.4.0
[5.3.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.2.3...v5.3.0
[5.2.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.2.2...v5.2.3
[5.2.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.2.1...v5.2.2
[5.2.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.2.0...v5.2.1
[5.2.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.1.1...v5.2.0
[5.1.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.1.0...v5.1.1
[5.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.0.3...v5.1.0
[5.0.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.0.2...v5.0.3
[5.0.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.0.1...v5.0.2
[5.0.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.0.0...v5.0.1
[5.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v4.1.0...v5.0.0
[4.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v4.0.0...v4.1.0
[4.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.4.0...v4.0.0
[3.4.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.3.1...v3.4.0
[3.3.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.3.0...v3.3.1
[3.3.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.2.0...v3.3.0
[3.2.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.1.1...v3.2.0
[3.1.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.0.3...v3.1.0
[3.0.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.0.2...v3.0.3
[3.0.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.0.1...v3.0.2
[3.0.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v3.0.0...v3.0.1
[3.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.12.0...v3.0.0
[2.12.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.11.0...v2.12.0
[2.11.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.9.1...v2.10.0
[2.9.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.9.0...v2.9.1
[2.9.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.8.4...v2.9.0
[2.8.4]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.8.3...v2.8.4
[2.8.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.8.2...v2.8.3
[2.8.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.8.1...v2.8.2
[2.8.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.8.0...v2.8.1
[2.8.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.7.1...v2.8.0
[2.7.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.7.0...v2.7.1
[2.7.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.6.1...v2.7.0
[2.6.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.6.0...v2.6.1
[2.6.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.5.0...v2.6.0
[2.5.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.4.2...v2.5.0
[2.4.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.4.1...v2.4.2
[2.4.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.4.0...v2.4.1
[2.4.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.3.0...v2.4.0
[2.3.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.2.0...v2.3.0
[2.2.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.5...v2.2.0
[2.1.5]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.4...v2.1.5
[2.1.4]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.3...v2.1.4
[2.1.3]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.2...v2.1.3
[2.1.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/2.0.0...v2.1.0
[2.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/1.0.0...2.0.0
[1.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/0.1.0...1.0.0
[0.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/releases/tag/0.1.0
