# Changelog
All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## 0.1.0 (Unreleased)

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
 - Go verison to 1.14 and against Go master branch in Travis CI
 - Travis CI file execute to execute website-test covering provider documentation
 - Provider uses Terraform Plugin SDK V2
 - resource_upcloud_server expanded with new functionality from UpCloud API 1.3
 - resource_upcloud_storage expaned with new functionalty from UpCloud API 1.3
 - resource_upclopud_tag expanded to implement read function
 
 
### Removed
 - Removed storage HCL blocks that failed due to referencing older UpCloud template ID
 - Removed the plan, price, price_zone and timezone UpCloud resources
 - resource_upcloud_ip removed and replaced by resource_upcloud_floating_ip_address
 - resource_upcloud_firewall_rule removed and replaced by resource_upcloud_firewall_rules
 - resource_upcloud_zone removed and replaced by zone and zones datasources
