# Examples for terraform-provider-upcloud

This is a full example which shows how you can set up your own UpCloud instance by using Terraform.

## Initializing local environment

Copy the `server.tf` example and run `terraform init` in the working directory.

```
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Using previously-installed registry.upcloud.com/upcloud/upcloud v0.1.0

The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, we recommend adding version constraints in a required_providers block
in your configuration, with the constraint strings suggested below.

* registry.upcloud.com/upcloud/upcloud: version = "~> 0.1.0"

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

## Configuring the plan

Set up your credentials to `server.tf`. Modify `login` block accordingly and set up your own SSH keys.
In this example you can also remove `keys` parameter, since UpCloud will auto generate a password for you
and deliver it via SMS.

Plan your changes.

```
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # upcloud_server.ubuntu will be created
  + resource "upcloud_server" "ubuntu" {
      + cpu      = (known after apply)
      + firewall = false
      + hostname = "ubuntu.example.tld"
      + id       = (known after apply)
      + mem      = (known after apply)
      + plan     = "1xCPU-1GB"
      + title    = (known after apply)
      + zone     = "de-fra1"

      + login {
          + create_password   = true
          + keys              = [
              + "ssh-rsa ...",
            ]
          + password_delivery = "sms"
          + user              = "terraform"
        }

      + network_interface {
          + bootable            = false
          + ip_address          = (known after apply)
          + ip_address_family   = "IPv4"
          + ip_address_floating = (known after apply)
          + mac_address         = (known after apply)
          + network             = (known after apply)
          + source_ip_filtering = true
          + type                = "public"
        }
      + network_interface {
          + bootable            = false
          + ip_address          = (known after apply)
          + ip_address_family   = "IPv4"
          + ip_address_floating = (known after apply)
          + mac_address         = (known after apply)
          + network             = (known after apply)
          + source_ip_filtering = true
          + type                = "utility"
        }

      + storage_devices {
          + address = (known after apply)
          + storage = (known after apply)
          + type    = (known after apply)
        }

      + template {
          + address = (known after apply)
          + id      = (known after apply)
          + size    = 25
          + storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
          + tier    = (known after apply)
          + title   = (known after apply)

          + backup_rule {
              + interval  = "daily"
              + retention = 8
              + time      = "0100"
            }
        }
    }

  # upcloud_storage.datastorage will be created
  + resource "upcloud_storage" "datastorage" {
      + id    = (known after apply)
      + size  = 10
      + tier  = "maxiops"
      + title = "/data"
      + zone  = "de-fra1"
    }

Plan: 2 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

This plan was saved to: plan

To perform exactly these actions, run the following command to apply:
    terraform apply "plan"
```

## Applying the plan

Apply the plan by running `terraform apply "plan"`.

```
upcloud_storage.datastorage: Creating...
upcloud_storage.datastorage: Still creating... [10s elapsed]
upcloud_storage.datastorage: Creation complete after 13s [id=0160a0c8-cb25-4726-85ef-d682439ca6b0]
upcloud_server.ubuntu: Creating...
upcloud_server.ubuntu: Still creating... [10s elapsed]
upcloud_server.ubuntu: Still creating... [20s elapsed]
upcloud_server.ubuntu: Still creating... [30s elapsed]
upcloud_server.ubuntu: Creation complete after 32s [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

The state of your infrastructure has been saved to the path
below. This state is required to modify and destroy your
infrastructure, so keep it safe. To inspect the complete state
use the `terraform show` command.

State path: terraform.tfstate

Outputs:

Public_ip = <SOME_IP_ADDRESS>
```

You can then log in to the server by running `ssh terraform@<SOME_IP_ADDRESS>`. Check your SMS messages if you didn't specify any SSH keys. You can print out the server details at any point by running the following
command:

```
$ terraform state show upcloud_server.example
```

## Modifying the plan

Next we will upgrade the plan to the next tier.
Do this by uncommenting relevant lines from the Terraform config file and run `terraform plan` again.


```
$ terraform plan -out plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

upcloud_storage.datastorage: Refreshing state... [id=0160a0c8-cb25-4726-85ef-d682439ca6b0]
upcloud_server.ubuntu: Refreshing state... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1]

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # upcloud_server.ubuntu will be updated in-place
  ~ resource "upcloud_server" "ubuntu" {
        cpu      = 1
        firewall = false
        hostname = "ubuntu.example.tld"
        id       = "0070bebf-5999-4e0f-b04f-459eb9d30ee1"
        mem      = 1024
      ~ plan     = "1xCPU-1GB" -> "1xCPU-2GB"
        title    = "ubuntu.example.tld (managed by terraform)"
        zone     = "de-fra1"

        login {
            create_password   = true
            keys              = [
                "ssh-rsa ...",
            ]
            password_delivery = "sms"
            user              = "terraform"
        }

        network_interface {
            bootable            = false
            ip_address          = "<SOME_IP_ADDRESS>"
            ip_address_family   = "IPv4"
            ip_address_floating = false
            mac_address         = "be:6d:ce:91:35:70"
            network             = "034a0abb-5b87-453f-bddc-f93863384e1f"
            source_ip_filtering = true
            type                = "public"
        }
        network_interface {
            bootable            = false
            ip_address          = "10.4.12.244"
            ip_address_family   = "IPv4"
            ip_address_floating = false
            mac_address         = "be:6d:ce:91:c0:a9"
            network             = "03406fbd-b9ce-48f8-b43b-2daf57ac5422"
            source_ip_filtering = true
            type                = "utility"
        }

        storage_devices {
            address = "virtio:1"
            storage = "0160a0c8-cb25-4726-85ef-d682439ca6b0"
            type    = "disk"
        }

        template {
            address = "virtio:0"
            id      = "01ae8e5d-85db-48ae-8c2f-844648911a4e"
            size    = 25
            storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
            tier    = "maxiops"
            title   = "terraform-ubuntu.example.tld-disk"

            backup_rule {
                interval  = "daily"
                retention = 8
                time      = "0100"
            }
        }
    }

Plan: 0 to add, 1 to change, 0 to destroy.

------------------------------------------------------------------------

This plan was saved to: plan

To perform exactly these actions, run the following command to apply:
    terraform apply "plan"
```

Then `terraform apply "plan"`.

```
$ terraform apply "plan"
upcloud_server.ubuntu: Modifying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1]
upcloud_server.ubuntu: Still modifying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1, 10s elapsed]
upcloud_server.ubuntu: Still modifying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1, 20s elapsed]
upcloud_server.ubuntu: Still modifying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1, 30s elapsed]
upcloud_server.ubuntu: Still modifying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1, 40s elapsed]
upcloud_server.ubuntu: Modifications complete after 40s [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1]

Apply complete! Resources: 0 added, 1 changed, 0 destroyed.

The state of your infrastructure has been saved to the path
below. This state is required to modify and destroy your
infrastructure, so keep it safe. To inspect the complete state
use the `terraform show` command.

State path: terraform.tfstate

Outputs:

Public_ip = <SOME_IP_ADDRESS>
```

Again, you may log in to the server and verify that memory has been increased.


## Cleaning up the environment

You can then destroy the instance by running `terraform destroy`. NOTE: You will lose all data.

```
$ terraform destroy
upcloud_storage.datastorage: Refreshing state... [id=0160a0c8-cb25-4726-85ef-d682439ca6b0]
upcloud_server.ubuntu: Refreshing state... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # upcloud_server.ubuntu will be destroyed
  - resource "upcloud_server" "ubuntu" {
      - cpu      = 1 -> null
      - firewall = false -> null
      - hostname = "ubuntu.example.tld" -> null
      - id       = "0070bebf-5999-4e0f-b04f-459eb9d30ee1" -> null
      - mem      = 2048 -> null
      - plan     = "1xCPU-2GB" -> null
      - title    = "ubuntu.example.tld (managed by terraform)" -> null
      - zone     = "de-fra1" -> null

      - login {
          - create_password   = true -> null
          - keys              = [
              - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDgYAlRed3dwPz5ffTpXNyD94zwdatYP+uxSqWNrR4OssZ0q7MksJhvvQqigr0IxP8sQLKA/t/zz/p403iza5nvX662rr4Xet2aoX5umfcTSvtPYuL7lTNo8LjN6jyIyykMD92nMxVmU9f8MpHrVVven5KJkcdaoQsDyjekZw2wmenUWabP0L9l5pKkVnu3RBpju9UuSEQIW5eaJXBLSAg6CWay2yAnCcmIKqf31r7WPnqY4Hq8IfzkarrITD5lE24Kdbq2KAF54dxlEyH8XeaS1IOX17VaNvgvjXy2YLECh4K3cfoRQKE1SB4tIbe7GL3purpVe4RIuDHyjTdUdwLcAiFIkpuKkmuihP9HqkUL9wrJlmEGa19A8JCPI+bw4+EXKTLrJjS2NUHy7ax6Ktlg5KxPqthwT+cCV+hAtfSf1gqgpAX0afjfkkGVPXwCHGdqEh203I3FfMpO/yITui+4ZUgnBvCZxYMUAIcTW9eG1A0Zcf3WvAL1aEdsSBv60qShzVmX86ZdsHJxgY37zgS00OXEycQvMrn+ZFNnYYOm/Nd/ewXOJ5wzV0GAFEGm5YcEG734+Pha3sjpJl1uEy7NW4rjMSHwwTgT2tEjsEUMQCPZAns17YJW+O6PaJ/N4MBHcT1z/w2gGZKLP03ZJytK3jlR2+mPQAUF5/iyNur1VQ== hallasmaa.touko@gmail.com",
            ] -> null
          - password_delivery = "sms" -> null
          - user              = "terraform" -> null
        }

      - network_interface {
          - bootable            = false -> null
          - ip_address          = "94.237.93.223" -> null
          - ip_address_family   = "IPv4" -> null
          - ip_address_floating = false -> null
          - mac_address         = "be:6d:ce:91:35:70" -> null
          - network             = "034a0abb-5b87-453f-bddc-f93863384e1f" -> null
          - source_ip_filtering = true -> null
          - type                = "public" -> null
        }
      - network_interface {
          - bootable            = false -> null
          - ip_address          = "10.4.12.244" -> null
          - ip_address_family   = "IPv4" -> null
          - ip_address_floating = false -> null
          - mac_address         = "be:6d:ce:91:c0:a9" -> null
          - network             = "03406fbd-b9ce-48f8-b43b-2daf57ac5422" -> null
          - source_ip_filtering = true -> null
          - type                = "utility" -> null
        }

      - storage_devices {
          - address = "virtio:1" -> null
          - storage = "0160a0c8-cb25-4726-85ef-d682439ca6b0" -> null
          - type    = "disk" -> null
        }

      - template {
          - address = "virtio:0" -> null
          - id      = "01ae8e5d-85db-48ae-8c2f-844648911a4e" -> null
          - size    = 25 -> null
          - storage = "Ubuntu Server 20.04 LTS (Focal Fossa)" -> null
          - tier    = "maxiops" -> null
          - title   = "terraform-ubuntu.example.tld-disk" -> null

          - backup_rule {
              - interval  = "daily" -> null
              - retention = 8 -> null
              - time      = "0100" -> null
            }
        }
    }

  # upcloud_storage.datastorage will be destroyed
  - resource "upcloud_storage" "datastorage" {
      - id    = "0160a0c8-cb25-4726-85ef-d682439ca6b0" -> null
      - size  = 10 -> null
      - tier  = "maxiops" -> null
      - title = "/data" -> null
      - zone  = "de-fra1" -> null
    }

Plan: 0 to add, 0 to change, 2 to destroy.

Changes to Outputs:
  - Public_ip = "94.237.93.223" -> null

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

upcloud_server.ubuntu: Destroying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1]
upcloud_server.ubuntu: Still destroying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1, 10s elapsed]
upcloud_server.ubuntu: Still destroying... [id=0070bebf-5999-4e0f-b04f-459eb9d30ee1, 20s elapsed]
upcloud_server.ubuntu: Destruction complete after 20s
upcloud_storage.datastorage: Destroying... [id=0160a0c8-cb25-4726-85ef-d682439ca6b0]
upcloud_storage.datastorage: Destruction complete after 0s

Destroy complete! Resources: 2 destroyed.
```

You can then verify that `terraform.tfstate` does not include the test server anymore.
Also, you can log in to UpCloud control panel and see that the instance and all disk resources created
by Terraform have been deleted.
