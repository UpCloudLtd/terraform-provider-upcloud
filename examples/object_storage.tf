# configure the provider
provider "upcloud" {
  # Your UpCloud credentials are read from the environment variables:
  # export UPCLOUD_USERNAME="Username of your UpCloud API user"
  # export UPCLOUD_PASSWORD="Password of your UpCloud API user"
}

# create an object storage instance with 2 buckets called "products" and "images"
resource "upcloud_object_storage" "my_object_storage" {
  # allocate up to 250GB of storage
  size = 250
  # the instance name, will form part of the url used to access the storage
  # instance so must conform to host naming rules.
  name = "storage-name"
  # The zone in which to create the instance
  zone = "fi-hel2"
  # The access key/username used to access the storage instance
  access_key = "admin"
  # The secret key/password used to access the storage instance
  secret_key  = "changeme"
  description = "catalogue"

  # Create a bucket called "products"
  bucket {
    name = "products"
  }

  # Create a bucket called "images"
  bucket {
    name = "images"
  }
}
