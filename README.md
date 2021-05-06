<a href="https://terraform.io">
  <img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for UpCloud

[![UpCloud Terraform provider tests](https://github.com/UpCloudLtd/terraform-provider-upcloud/workflows/UpCloud%20Terraform%20provider%20tests/badge.svg)](https://github.com/UpCloudLtd/terraform-provider-upcloud/actions)

This provider is developed by UpCloud, contributions from the community are welcomed!

* Check Github issues or create more issues
* Check `examples/` directory for examples and test them
* Improve documentation
* Website: https://www.terraform.io
* [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
* Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

This Terraform provider is a plugin for Terraform which provides capabilities to manage your UpCloud products such as servers, storages, networks and IP addresses.

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 0.15.x or later

## Quick Start

Install [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli).

Set environment variables for authentication (on command line, in .bashrc, .zshrc, ...):

```sh
export UPCLOUD_USERNAME="upcloud-api-access-enabled-user"
export UPCLOUD_PASSWORD="verysecretpassword"
```

Create a example_create_server.tf file:

```terraform
# Configure the UpCloud provider
provider "upcloud" {}

# Create a server
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  # Set the operating system
  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"

    # Use the size allotted by the 1xCPU-1GB plan
    size = 25
  }

  # Add a public IP address
  network_interface {
    type = "public"
  }
}
```

Run Terraform with e.g. `terraform apply`.

## Using the provider

Before using the provider, install [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli).

You need to set UpCloud credentials in shell environment variable (.bashrc, .zshrc or similar) to be able to use the provider:

* `export UPCLOUD_USERNAME="Username for Upcloud API user"` - Your API access enabled user's username
* `export UPCLOUD_PASSWORD="Password for Upcloud API user"` - Your API access enabled user's password

To allow API access to your UpCloud account, you need to allow API connections by visiting [Account-page](https://hub.upcloud.com/account) in your UpCloud Hub.

We recommend you to set up a sub-account specifically for the API usage with its own username and password, as it allows you to assign specific permissions for increased security:

1. Open the [People-page](https://hub.upcloud.com/people) in the UpCloud Hub
2. Click **Add** in top-right corner and fill in the required details, and check the **Allow API connections** checkbox to enable API for the sub-account. You can also limit the API connections to a specific IP address or address range for additional security
3. Click the **Create subaccount** button at the bottom left of the page to create the sub-account

Below is an example configuration on how to create a server using the Terraform provider with Terraform 0.13 or later:

```terraform
# set the provider version
terraform {
  required_providers {
    upcloud = {
      source = "UpCloudLtd/upcloud"
      version = "~> 2.0"
    }
  }
}

# configure the provider
provider "upcloud" {
  # Your UpCloud credentials are read from the environment variables:
  # export UPCLOUD_USERNAME="Username of your UpCloud API user"
  # export UPCLOUD_PASSWORD="Password of your UpCloud API user"
}

# create a server
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  # Declare network interfaces
  network_interface {
    type = "public"
  }

  network_interface {
    type = "utility"
  }

  # Include at least one public SSH key
  login {
    user = "terraform"
    keys = [
      "<YOUR SSH PUBLIC KEY>",
    ]
    create_password = false
  }

  # Provision the server with Ubuntu
  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"

    # Use all the space allotted by the selected simple plan
    size = 25

    # Enable backups
    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }
}
```

Terraform 0.12 or earlier:

```terraform
# configure the provider
provider "upcloud" {
  # Your UpCloud credentials are read from the environment variables:
  # export UPCLOUD_USERNAME="Username of your UpCloud API user"
  # export UPCLOUD_PASSWORD="Password of your UpCloud API user"
  version = "~> 2.0"
}

# create a server
resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  # Declare network interfaces
  network_interface {
    type = "public"
  }

  network_interface {
    type = "utility"
  }

  # Include at least one public SSH key
  login {
    user = "terraform"
    keys = [
      "<YOUR SSH PUBLIC KEY>",
    ]
    create_password = false
  }

  # Provision the server with Ubuntu
  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"

    # Use all the space allotted by the selected simple plan
    size = 25

    # Enable backups
    backup_rule {
      interval  = "daily"
      time      = "0100"
      retention = 8
    }
  }
}
```

For more examples, check the `examples/` directory or visit [official Terraform documentation](https://registry.terraform.io/providers/UpCloudLtd/upcloud/latest/docs).

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is _required_).

Get the provider source code:

```sh
$ git clone git@github.com:UpCloudLtd/terraform-provider-upcloud.git
$ cd terraform-provider-upcloud
```

To compile the provider, run `go build`. This will build the provider and put the provider binary in the current directory.

```sh
$ go build
```

In the majority of cases the `make` command will be executed to allow the provider binary to be discovered by Terraform.

```sh
$ make
```

Symlink the provider into a directory for Terraform to find it:

```sh
$ mkdir -p $HOME/.terraform.d/plugins
$ ln -s $GOBIN/terraform-provider-upcloud $HOME/.terraform.d/plugins
```

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

### Consuming local provider with Terraform 0.13.0

With the release of Terraform 0.13.0 the discovery of a locally built provider binary has changed.
These changes have been made to allow all providers to be discovered from public and provider registries.

The UpCloud makefile has been updated with a new target to build the provider binary into the right location for discovery.
The following commands will allow you to build and execute terraform with the provider locally.

Update your terraform files with the following terraform configuration block.  A standard name for a file with the following HCL is `version.tf`.

```terraform
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

### Developing in Docker

You can also develop/build/test in Docker. After you've cloned the repository:

Create a docker container with golang as base:

```sh
docker run -it -v `pwd`:/work -w /work golang bash
```

Install Terraform:

```sh
cd /tmp
git clone https://github.com/hashicorp/terraform.git
cd terraform
go install
```

Build the UpCloud provider:

```sh
cd /work
make build_0_13
```

Run Terraform files, e.g. the examples:

```sh
cd /tmp
cp /work/examples/01_server.tf .
export UPCLOUD_USERNAME="upcloud-api-access-enabled-user"
export UPCLOUD_PASSWORD="verysecretpassword"
terraform init
terraform apply
```

After exiting the container, you can connect back to the container:

```sh
docker start -ai <container ID here>
```
