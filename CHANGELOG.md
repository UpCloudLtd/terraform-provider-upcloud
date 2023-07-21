# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

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
- New [examples](../blob/master/examples) of using the provider
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

[Unreleased]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.12.0...HEAD
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
