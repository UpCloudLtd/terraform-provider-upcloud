# Changelog

All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## [Unreleased]

### Added

- Added title field to the server resource

### Fixed

- server: fix custom plan updates (cpu/mem)

### Changed

- server: new hostname validator

## [2.1.2]

### Added

- Added simple backups support (#188)

### Fixed

- Prevent empty tags from replanning a server (#178)
- Make sure either storage devices or template are required on the server resource

## [2.1.1]

### Fixed

- fix(client): fix user-agent value (#165)

## [2.1.0]

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

## [2.0.0]

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

## [1.0.0]

Updated upcloud-go-api, added build/CI scripts, and repackaged 0.1.0 as 1.0.0.

## [0.1.0]

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

[Unreleased]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.2...HEAD
[2.1.2]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/2.0.0...v2.1.0
[2.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/1.0.0...2.0.0
[1.0.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/compare/0.1.0...1.0.0
[0.1.0]: https://github.com/UpCloudLtd/terraform-provider-upcloud/releases/tag/0.1.0
