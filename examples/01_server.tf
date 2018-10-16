provider "upcloud" {
  # You need to set UpCloud credentials in shell environment variable
  # using .bashrc, .zshrc or similar
  # export UPCLOUD_USERNAME="Username for Upcloud API user"
  # export UPCLOUD_PASSWORD="Password for Upcloud API user"
}

resource "upcloud_server" "test" {
  count    = "1"
  zone     = "fi-hel1"
  hostname = "nikoh-tf-provider-testing-${format("%04d", count.index + 1)}"

  cpu      = "2"
  mem      = "1024"

  # Login details
  login {
    user = "tf"
    keys = [
      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCYn6VuEgiH3//qpSa/b3Khrjy3Z4Q4fvvhRNRDrZaJqddLvQLCtoL2ktoke7+0jTcR4Vydi8bk8csUQlZxpWC6SIfif+tB8HjwusbUfLT5I5fJEI/O7gtktvtWkK4GnePFXYIdgKlXKRJ92xFnNOGV.............",
    ]

    create_password   = true
    password_delivery = "email"
  }
}


resource "upcloud_storage" "my-storage" {
			size  = 10
			tier  = "maxiops"
			title = "My data collection"
			zone  = "fi-hel1"
      templatize = true
      source_storage_id = "01000000-0000-4000-8000-000020040100"
      instances = ["${upcloud_server.test.*.id}"]
		}

output "ipv4_addresses" {
  value = ["${upcloud_server.test.*.ipv4_address}"]
}

output "storage_uuids" {
  value = ["${upcloud_storage.my-storage.*.id}"]
}
