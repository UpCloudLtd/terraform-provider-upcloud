# Terraform Provider

This provider is currently under active development. It is not production-ready yet so you are advised to chime in and help!

* Check Github issues or create more issues
* Check `examples/` directory for examples and test them
* Improve documentation

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
