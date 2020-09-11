# Examples for terraform-provider-upcloud

This is a full example which shows how you can set up your own UpCloud instance by using Terraform.

## Initializing local environment

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

Clone this example and init Terraform in the example folder:

```sh
$ cd terraform-upcloud-provider/examples
$ terraform init

Initializing the backend...

Initializing provider plugins...

Terraform has been successfully initialized!
...
```

## Configuring the plan

Set up your credentials to `01_server.tf`. Modify `login` block accordingly and set up your own SSH keys.
In this example you can also remove `keys` parameter, since UpCloud will auto generate a password for you
and deliver it via SMS.

Plan your changes.

```
$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + upcloud_server.test
      id:                                <computed>
      cpu:                               <computed>
      hostname:                          "ubuntu.example.tld"
      ipv4:                              "true"
      ipv4_address:                      <computed>
      ipv4_address_private:              <computed>
      ipv6:                              "true"
      ipv6_address:                      <computed>
      login.#:                           "1"
      login.123878037.create_password:   "true"
      login.123878037.keys.#:            "1"
      login.123878037.keys.0:            "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCYn6VuEgiH3//qpSa/b3Khrjy3Z4Q4fvvhRNRDrZaJqddLvQLCtoL2ktoke7+0jTcR4Vydi8bk8csUQlZxpWC6SIfif+tB8HjwusbUfLT5I5fJEI/O7gtktvtWkK4GnePFXYIdgKlXKRJ92xFnNOGV+el2zug78QahsrzsyV0Cucfjb7twPyojh5iPl3gf6f7NBHVnsqNELhJqmpo4uY+vSTfHx0siyIGP0U/Jz9dB64kbnoG6GL2fh3CEQ950Ll2luY/cfX52SO+WX/nl156A2VVCozkOSE3wbZ501Gd1508KY7ctuaqOue4DF8ZuQ1uzv4Lf9sfg4Bv4jBMTu4tvB"
      login.123878037.password_delivery: "sms"
      login.123878037.user:              "tf"
      private_networking:                "true"
      storage_devices.#:                 "3"
      storage_devices.0.action:          "clone"
      storage_devices.0.address:         <computed>
      storage_devices.0.id:              <computed>
      storage_devices.0.size:            "50"
      storage_devices.0.storage:         "Ubuntu Server 16.04 LTS (Xenial Xerus)"
      storage_devices.0.title:           <computed>
      storage_devices.1.action:          "attach"
      storage_devices.1.address:         <computed>
      storage_devices.1.id:              <computed>
      storage_devices.1.size:            "-1"
      storage_devices.1.storage:         "01000000-0000-4000-8000-000020010301"
      storage_devices.1.title:           <computed>
      storage_devices.1.type:            "cdrom"
      storage_devices.2.action:          "create"
      storage_devices.2.address:         <computed>
      storage_devices.2.id:              <computed>
      storage_devices.2.size:            "25"
      storage_devices.2.tier:            "maxiops"
      storage_devices.2.title:           <computed>
      title:                             <computed>
      zone:                              "fi-hel1"

Plan: 1 to add, 0 to change, 0 to destroy.
```

## Applying the plan

Apply the plan by running `terraform apply`.

```
$ terraform apply
upcloud_server.test: Creating...
  cpu:                               "" => "<computed>"
  hostname:                          "" => "ubuntu.example.tld"
  ipv4:                              "" => "true"
  ipv4_address:                      "" => "<computed>"
  ipv4_address_private:              "" => "<computed>"
  ipv6:                              "" => "true"
  ipv6_address:                      "" => "<computed>"
  login.#:                           "" => "1"
  login.123878037.create_password:   "" => "true"
  login.123878037.keys.#:            "" => "1"
  login.123878037.keys.0:            "" => "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCYn6VuEgiH3//qpSa/b3Khrjy3Z4Q4fvvhRNRDrZaJqddLvQLCtoL2ktoke7+0jTcR4Vydi8bk8csUQlZxpWC6SIfif+tB8HjwusbUfLT5I5fJEI/O7gtktvtWkK4GnePFXYIdgKlXKRJ92xFnNOGV+el2zug78QahsrzsyV0Cucfjb7twPyojh5iPl3gf6f7NBHVnsqNELhJqmpo4uY+vSTfHx0siyIGP0U/Jz9dB64kbnoG6GL2fh3CEQ950Ll2luY/cfX52SO+WX/nl156A2VVCozkOSE3wbZ501Gd1508KY7ctuaqOue4DF8ZuQ1uzv4Lf9sfg4Bv4jBMTu4tvB"
  login.123878037.password_delivery: "" => "sms"
  login.123878037.user:              "" => "tf"
  private_networking:                "" => "true"
  storage_devices.#:                 "" => "3"
  storage_devices.0.action:          "" => "clone"
  storage_devices.0.address:         "" => "<computed>"
  storage_devices.0.id:              "" => "<computed>"
  storage_devices.0.size:            "" => "50"
  storage_devices.0.storage:         "" => "Ubuntu Server 16.04 LTS (Xenial Xerus)"
  storage_devices.0.title:           "" => "<computed>"
  storage_devices.1.action:          "" => "attach"
  storage_devices.1.address:         "" => "<computed>"
  storage_devices.1.id:              "" => "<computed>"
  storage_devices.1.size:            "" => "-1"
  storage_devices.1.storage:         "" => "01000000-0000-4000-8000-000020010301"
  storage_devices.1.title:           "" => "<computed>"
  storage_devices.1.type:            "" => "cdrom"
  storage_devices.2.action:          "" => "create"
  storage_devices.2.address:         "" => "<computed>"
  storage_devices.2.id:              "" => "<computed>"
  storage_devices.2.size:            "" => "25"
  storage_devices.2.tier:            "" => "maxiops"
  storage_devices.2.title:           "" => "<computed>"
  title:                             "" => "<computed>"
  zone:                              "" => "fi-hel1"
upcloud_server.test: Still creating... (10s elapsed)
upcloud_server.test: Still creating... (20s elapsed)
upcloud_server.test: Still creating... (30s elapsed)
upcloud_server.test: Still creating... (40s elapsed)
upcloud_server.test: Creation complete after 46s (ID: <snip>)

Outputs:

ip = <SOME IP ADDRESS>
```

You can then log in to the server by running `ssh tf@<SOME IP ADDRESS>`. Check your SMS messages if you didn't specify any SSH keys. You can print out the server details at any point by running the following
command:

```
$ terraform state show upcloud_server.test
```

## Modifying the plan

Next we increase the CPU amount from 1 (default) to 2 and memory amount from 512 MB (default) to 1024 MB.
Do this by uncommenting relevant lines from the Terraform config file and run `terraform plan` again.


```
$ terraform plan
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  ~ upcloud_server.test
      cpu: "1" => "2"
      mem: "512" => "1024"
```

Then `terraform apply`.

```
$ terraform apply
upcloud_server.test: Refreshing state... (ID: <snip>)
upcloud_server.test: Modifying... (ID: <snip>)
  cpu: "1" => "2"
  mem: "512" => "1024"
upcloud_server.test: Still modifying... (ID: <snip>, 10s elapsed)
upcloud_server.test: Still modifying... (ID: <snip>, 20s elapsed)
upcloud_server.test: Still modifying... (ID: <snip>, 30s elapsed)
upcloud_server.test: Still modifying... (ID: <snip>, 40s elapsed)
upcloud_server.test: Modifications complete after 40s (ID: <snip>)

Outputs:

ipv4_address = <SOME IP ADDRESS>
```

Again, log in to the server and verify that memory has been increased.


## Cleaning up the environment

You can then destroy the instance by running `terraform destroy`. NOTE: You will lose all data.

```
$ terraform destroy
Terraform will perform the following actions:

  - upcloud_server.test

Plan: 0 to add, 0 to change, 1 to destroy.

Do you really want to destroy?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

upcloud_server.test: Destroying... (ID: <snip>)
upcloud_server.test: Still destroying... (ID: <snip>, 10s elapsed)
upcloud_server.test: Destruction complete after 10s

Destroy complete! Resources: 1 destroyed.
```

You can then verify that `terraform.tfstate` does not include the test server anymore.
Also, you can log in to UpCloud control panel and see that the instance and all disk resources created
by Terraform have been deleted.
