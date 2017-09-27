# Examples for terraform-provider-upcloud

First install this provider and initialize Terraform in this directory.

```
$ go install github.com/vtorhonen/terraform-upcloud-provider
$ terraform init

Initializing provider plugins...

Terraform has been successfully initialized!
```

Plan your changes.

```
$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + upcloud_server.test
      id:       <computed>
      hostname: "my-awesome-hostname"
      title:    "my awesome title"
      zone:     "fi-hel1"


Plan: 1 to add, 0 to change, 0 to destroy.
```

Then apply the plan.

```
$ upcloud_server.test: Creating...
  hostname: "" => "my-awesome-hostname"
  title:    "" => "my awesome title"
  zone:     "" => "fi-hel1"
upcloud_server.test: Still creating... (10s elapsed)
upcloud_server.test: Still creating... (20s elapsed)
upcloud_server.test: Still creating... (30s elapsed)
upcloud_server.test: Still creating... (40s elapsed)
upcloud_server.test: Still creating... (50s elapsed)
upcloud_server.test: Creation complete after 51s (ID: <snip>)
```
