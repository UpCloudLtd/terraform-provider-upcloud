# Terraform Provider

This provider is developed by UpCloud, contributions from the community are welcomed!

* Check Github issues or create more issues
* Check `examples/` directory for examples and test them
* Improve documentation

* Website: https://www.terraform.io
* [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
* Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 0.12.x, (to execute the provider plugin)
* [Go](https://golang.org/doc/install) 1.14.x or greater, (to build the provider plugin)

## Building The Provider

Get and install the provider:

```sh
$ git clone git@github.com:UpCloudLtd/terraform-provider-upcloud.git
$ cd terraform-provider-upcloud
```

Build and symlink the provider into a folder (also make sure it exists) where Terraform looks for it:

```sh
$ make
$ mkdir -p $HOME/.terraform.d/plugins
$ ln -s $GOBIN/terraform-provider-upcloud $HOME/.terraform.d/plugins
```

## Using the provider

You need to set UpCloud credentials in shell environment variable (.bashrc, .zshrc or similar) to be able to use the provider:

* `export UPCLOUD_USERNAME="Username for Upcloud API user"` - Your API access enabled users username
* `export UPCLOUD_PASSWORD="Password for Upcloud API user"` - Your API access enabled users password

To allow API access to your UpCloud account, you first need to enable the API permissions by visiting [My Account -> User accounts](https://my.upcloud.com/account) in your UpCloud Control Panel. We recommend you to set up a sub-account specifically for the API usage with its own username and password, as it allows you to assign specific permissions for increased security.

Click **Add user** and fill in the required details, and check the “**Allow API connections**” checkbox to enable API for the user. You can also limit the API connections to a specific IP address or address range for additional security. Once you are done entering the user information, hit the **Save** button at the bottom of the page to create the new username.

For more instructions, check out examples folder.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is _required_).

To compile the provider, run `go build`. This will build the provider and put the provider binary in the current directory.

```sh
$ go build
```
In the majority of cases the ```make``` command will be executed to allow the provider binary to be discovered by Terraform.

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

In order to run an individual acceptance test, the '-run' flag can be used together with a regular expression.
The following example uses a regular expression matching single test called 'TestUpcloudServer_basic'.

```sh
$ make testacc TESTARGS='-run=TestUpcloudServer_basic'
```

The following example uses a regular expression to execute a grouping of basic acceptance tests.

```sh
$ make testacc TESTARGS='-run=TestUpcloudServer_*'
```

In order to view the provider documentation locally, you can run `make website`.
A docker container will start and a URl to the documentation will be returned.

```sh
$ make website

...
==> Starting upcloud provider website in Docker...
== The Middleman is loading
==
==> See upcloud docs at http://localhost:4567/docs/providers/upcloud
...
``` 

A website test can be execute to confirm that the links inside the website docs are not broken.
This test can be run through the following command

```sh
$ make website-test
```

## Consuming local provider with Terraform 0.13.0

With the release of Terraform 0.13.0 the discovery of a locally built provider binary has changed.
These changes have been made to allow all providers to be discovered from public and provider registries.

The UpCloud makefile has been updated with a new target to build the provider binary into the right location for discovery.
The following commands will allow you to build and execute terraform with the provider locally.

Update your terraform files with the following terraform configuration block.  A standard name for a file with the following HCL is `version.tf`.

```
terraform {
  required_providers {
    upcloud = {
      source = "registry.upcloud.com/upcloud/upcloud"
    }
  }
  required_version = ">= 0.13"
}
```

The following make command can be executed to build and place the provider in the correct directory location.

```sh
$ make build_0_13
```

The UpCloud provider will be built and placed in the following location under the `~/.terraform.d/plugins` directory.
The version number will match the value specified in the makefile and in this case the version is 0.1.0.

```
~/.terraform.d/plugins
└── registry.upcloud.com
    └── upcloud
        └── upcloud
            └── 0.1.0
                └── darwin_amd64
                    └── terraform-provider-upcloud_v0.1.0
``` 

After the provider has been built you can then use standard terraform commands can be executed as normal.