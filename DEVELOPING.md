# Developing the Provider
## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 0.12.x or later
* [Go](https://golang.org/doc/install) > 1.14

Get the provider source code:

```sh
git clone git@github.com:UpCloudLtd/terraform-provider-upcloud.git
cd terraform-provider-upcloud
```

To compile the provider, run `go build`. This will build the provider and put
the provider binary in the current directory.

```sh
go build
```

In the majority of cases the `make` command will be executed to build the
binary in the correct directory.

```sh
make
```

The UpCloud provider will be built and placed in the following location under
the `~/.terraform.d/plugins` directory.  The version number will match the
value specified in the makefile and in this case the version is 2.0.0.

```
~/.terraform.d/plugins
└── registry.upcloud.com
    └── upcloud
        └── upcloud
            └── 2.0.0
                └── darwin_amd64
                    └── terraform-provider-upcloud_v2.0.0
```

After the provider has been built you can then use standard terraform commands.

Use this provider config with the local version:

```
terraform {
  required_providers {
    upcloud = {
      source = "registry.upcloud.com/upcloud/upcloud"
      version = "~> 2.1"
    }
  }
}
```

## Update database properties

Prerequisites:

- Set your credentials to `UPCLOUD_USERNAME` and `UPCLOUD_PASSWORD` environment variables.
- Ensure you have `jq` and `upctl` installed.

Run `make generate` to update database properties schemas.

```sh
make generate
```

## Testing

To lint the providers source-code, run `golangci-lint run`. See [golangci-lint docs](https://golangci-lint.run/usage/install/) for installation instructions.

```sh
golangci-lint run
```

In order to test the provider, you can simply run `make test`.

```sh
make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```sh
make testacc
```

In order to run an individual acceptance test, the '-run' flag can be used
together with a regular expression.  The following example uses a regular
expression matching single test called 'TestUpcloudServer_basic'.

```sh
make testacc TESTARGS='-run=TestUpcloudServer_basic'
```

The following example uses a regular expression to execute a grouping of basic
acceptance tests.

```sh
make testacc TESTARGS='-run=TestUpcloudServer_*'
```

In order to view the documentation change rendering visite
[the terraform documentation preview](https://registry.terraform.io/tools/doc-preview).

## Generating documentation

The documentation in [docs](./docs/) directory is generated with [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs). This is done automatically when changes are merged to the `main` branch. If there are documentation changes, actions will create a new pull request for documentation changes. Review and merge this pull-request.

To generate the docs locally, run `make docs`. This installs the tool and builds the docs.

```sh
make docs
```

## Go version upgrades

Upgrading Go version for the project requires the following changes:
- Change Go version in [go.mod file](https://github.com/UpCloudLtd/terraform-provider-upcloud/blob/v2.4.1/go.mod)
- Change `go-version` argument for `Setup Go` step in [acceptance test GitHub action file](https://github.com/UpCloudLtd/terraform-provider-upcloud/blob/v2.4.1/.github/workflows/terraform.yml)
- Change `GO_VERSION` argument in [Dockerfile](https://github.com/UpCloudLtd/terraform-provider-upcloud/blob/v2.4.1/release/Dockerfile)

After updating those files, make sure that you can still build the project (`make build`) and that docker image builds without any errors (`docker image build ./release`).

Once that is done it should be safe to release a new version of the provider.

## Developing in Docker

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
make build
```

Run Terraform files, e.g. the examples:

```sh
cd /tmp
cp /work/examples/server.tf .
# change the provider source in server.tf to "registry.upcloud.com/upcloud/upcloud"
export UPCLOUD_USERNAME="upcloud-api-access-enabled-user"
export UPCLOUD_PASSWORD="verysecretpassword"
terraform init
terraform apply
```

After exiting the container, you can connect back to the container:

```sh
docker start -ai <container ID here>
```

## Debugging

UpCloud provider can be run in debug mode using a debugger such as Delve.  
For more information, see [Terraform docs](https://www.terraform.io/docs/extend/debugging.html#starting-a-provider-in-debug-mode)  

Environment variables `UPCLOUD_DEBUG_API_BASE_URL` and `UPCLOUD_DEBUG_SKIP_CERTIFICATE_VERIFY` can be used for HTTP client debugging purposes. More information can be found in the client's [README](https://github.com/UpCloudLtd/upcloud-go-api/blob/986ca6da9ca85ff51ecacc588215641e2e384cfa/README.md#debugging) file.

## Consuming local provider with Terraform 0.12.0

With the release of Terraform 0.13.0 the discovery of a locally built provider
binary has changed.  These changes have been made to allow all providers to be
discovered from public and provider registries.

The UpCloud makefile supports the old, Terraform 0.12 style where plugin
directory structure is not relevant and only the binary name matters.

The following make command can be executed to build and place the provider in
the Go binary directory. Make sure your PATH includes the Go binary directory.

```sh
make build_0_12
```

After the provider has been built and can be executed, you can then use
standard terraform commands as normal.

### Using local provider with Terraform 0.12 or earlier

```terraform
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