# For object storage import to work properly, you need to set environment variables for access and secret key.
# The environment variables names are UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_{name} and UPCLOUD_OBJECT_STORAGE_SECRET_KEY_{name}
# where {name} is the name of your object storage instance (not the resource label!), all uppercased, and with all dashes (-)
# replaced with underscores (_). So importing an object storage that is named "my-storage" will look like this:

UPCLOUD_OBJECT_STORAGE_ACCESS_KEY_MY_STORAGE=accesskey \
UPCLOUD_OBJECT_STORAGE_SECRET_KEY_MY_STORAGE=supersecret \
terraform import upcloud_object_storage.example_storage 06c1f4b6-faf2-47d0-8896-1d941092b009