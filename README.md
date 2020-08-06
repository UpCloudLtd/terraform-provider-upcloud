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
$ go mod download
$ go install
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