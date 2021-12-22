# Object storage instance called storage-name in the fi-hel2 zone, with 2 buckets called "products" and "images".
resource "upcloud_object_storage" "my_object_storage" {
  size        = 250            # allocate up to 250GB of storage
  name        = "storage-name" # the instance name, will form part of the url used to access the storage instance so must conform to host naming rules.
  zone        = "fi-hel2"      # The zone in wgich to create the instance
  access_key  = "admin"        # The access key/username used to access the storage instance
  secret_key  = "changeme"     # The secret key/password used to access the storage instance
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
