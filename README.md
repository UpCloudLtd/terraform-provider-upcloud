# Terraform Provider

This provider is currently under active development. It is not production-ready yet so you are advised to chime in and help!

* Check Github issues or create more issues
* Check `examples/` directory for examples and test them
* Improve documentation
* Improve functionality

* Website: https://www.terraform.io
* [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
* Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 0.10.x
* [Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)
* [Go dep](https://github.com/golang/dep) (to install vendor deps)

## Building The Provider

Get and install the provider:

```sh
$ mkdir -p $GOPATH/src/github.com/UpCloudLtd; cd $GOPATH/src/github.com/UpCloudLtd
$ git clone git@github.com:UpCloudLtd/terraform-provider-upcloud.git
$ cd terraform-provider-upcloud
$ dep ensure
```

Build and symlink the provider into a folder (also make sure it exists) where Terraform looks for it:

```sh
$ cd $GOPATH/src/github.com/UpCloudLtd/terraform-provider-upcloud
$ make build
$ mkdir -p $HOME/.terraform.d/plugins
$ ln -s $GOPATH/bin/terraform-provider-upcloud $HOME/.terraform.d/plugins/terraform-provider-upcloud
```

## Using the provider

You need to set UpCloud credentials in shell environment variable (.bashrc, .zshrc or similar) to be able to use the provider:

* `export UPCLOUD_USERNAME="Username for Upcloud API user"` - Your API access enabled users username
* `export UPCLOUD_PASSWORD="Password for Upcloud API user"` - Your API access enabled users password

To allow API access to your UpCloud account, you first need to enable the API permissions by visiting [My Account -> User accounts](https://my.upcloud.com/account) in your UpCloud Control Panel. We recommend you to set up a sub-account specifically for the API usage with its own username and password, as it allows you to assign specific permissions for increased security.

Click **Add user** and fill in the required details, and check the “**Allow API connections**” checkbox to enable API for the user. You can also limit the API connections to a specific IP address or address range for additional security. Once you are done entering the user information, hit the **Save** button at the bottom of the page to create the new username.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is _required_). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-upcloud
...
```

In order to test the provider, you can simply run `make test`.
Obs. This command runs only unit tests for the provider and acceptance tests will be skipped as a default if environment variable TF_ACC hasn't set

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

## Observations related on Upcloud API 
* API documentation https://www.upcloud.com/api/1.2.6/
* Check the server(instance) creation limitations https://www.upcloud.com/api/1.2.6/8-servers/ `SERVER_CREATING_LIMIT_REACHED`
  * Can be resolved by using the Terraform --parallelism flag https://www.terraform.io/docs/commands/apply.html#parallelism-n
* False positive instance status (maintenance) response can cause some unexpected functionality with the provider (false positive error messages)

## Improve provider
* Template update feature
  - [X]  Update a storage which has been templatized. Depedent on the storage access levels (private or public)
  - [ ] Write an acceptance test case for the template update (need a storage templatize support at first)
  - [X]  Fix timeout issues e.g. when an instance processing takes more time than 5 minutes

* Testing
  - [ ] Write unit tests (table) for the functions
  - [ ] Write more valid Terraform acceptance test cases for the resources

* Restructuring
  - [ ] Rename the upcloud_server resource into instance (more clear)
  - [ ] Move all storage functionalities from the upcloud_server resource into the storage resource
    * Instance requirements
      - [ ] Update function should be able to modify (zone, cpu, memory, ip addresses and hostname)
    * Storage requirements
      - [ ] Add templatize support for the storages (e.g. managed by the templatize flag)
      - [ ] Attach and deattach instances

* Use partial state in the resource update functions
  - [X] Instance (server)
  - [ ] Storage
  - [ ] Firewall
  - [ ] Tag
  - [ ] Plan

* Resource validation (All user inputs have to be validated at the Terraform plan phase)
  - [ ] Write validators for all the upcloud resources according to Upcloud API
* Check validators.go file and the Terraform validation help package in here https://github.com/hashicorp/terraform/blob/master/helper/validation/validation.go
