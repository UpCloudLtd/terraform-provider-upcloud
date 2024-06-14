# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

### Added

- upcloud_router: support `labels` field

### Fixed

- upcloud_network: detect if resource was deleted outside of Terraform
- upcloud_network_peering: detect if resource was deleted outside of Terraform
- upcloud_floating_ip_address: replace floating IP address, if `family` or `zone` have changes.

## [5.5.0] - 2024-06-04

### Added

- upcloud_zone: `parent_zone` field.
- upcloud_network: support `labels` field
- gateway: uuid field for `upcloud_gateway_connection` resource
- gateway: uuid field for `upcloud_gateway_connection_tunnel` resource

### Fixed

- upcloud_managed_database_*: do not include empty or unmodified nested properties values in the API request. I.e., filter nested properties similarly than main level properties.

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

- managed object storage: fix error when refreshing a state after the resource was deleted outside of Terraform; applies to all upcloud_managed_object_storage* resources

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

[Unreleased]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v5.5.0...HEAD
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
