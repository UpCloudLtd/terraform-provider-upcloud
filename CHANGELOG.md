# Changelog
All notable changes to this project will be documented in this file.
See updating [Changelog example here](https://keepachangelog.com/en/1.0.0/)

## 0.1.0 (Unreleased)

### Added
 - Changelog to highlight key alterations across future releases 
 - Website directory for future provider documentation
 - Vendor directory through go modules to cover CI builds

### Changed
 - README and examples/README to cover local builds, setup and test execution
 - Go verison to 1.14 and against Go master branch in Travis CI
 - Travis CI file execute to execute website-test covering provider documentation
 - Provider uses Terraform Plugin SDK V1
 
### Removed
 - Removed storage HCL blocks that failed due to referencing older UpCloud template ID
 - Removed the plan, price, price_zone and timezone UpCloud resources
 