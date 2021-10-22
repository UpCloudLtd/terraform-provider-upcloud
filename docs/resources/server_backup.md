# upcloud_server_backup (Resource)

The UpCloud `server_backup` resource allows managing the backup rules for all the storages attached to a specific server in simplified manner.

## Example

### Basic

The following example shows the creation of a `server` with additional `storage` and a `server_backup` resource. Both "addon" storage and the main server template storage will be backed up everyday at 11am, and the backups will be kept for 7 days 

```hcl
resource "upcloud_storage" "addon" {
  title = "addon"
  size = 10
  zone = "pl-waw1"
}

resource "upcloud_server" "example" {
  hostname = "terraform.example.tld"
  zone     = "de-fra1"
  plan     = "1xCPU-1GB"

  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size    = 25
  }

  network_interface {
    type = "public"
  }

  storage_devices {
    storage = upcloud_storage.addon.id
  }
}

resource "upcloud_server_backup" "example_backup" {
  server = upcloud_server.example.id
  plan = "dailies"
  time = "1100"
}
```

## Schema

### Required

- **server** (String) ID of a server that should be backed up
- **plan** (String) The backup plan. Can be one of "dailies", "weeklies" or "monthlies"
- **time** (String) The exact time at which the backup should be taken, in hhmm format.

## Notes

### Conflicting backup rules
Please note that it is impossible to use `storage_backup` and `server_backup` resources for the same resources. Trying to set `storage_backup` for a storage that is attached to a server that already has a `server_backup` will fail.

For example, applying the following configuration will fail:
```hcl
resource "upcloud_storage" "addon" {
  title = "addon"
  size = 10
  zone = "pl-waw1"
}

resource "upcloud_server" "example" {
  zone = "pl-waw1"
  plan = "1xCPU-1GB"
  hostname = "mainx1"
  
  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size = 10
  }

  network_interface {
    type = "public"
  }

  storage_devices {
    storage = upcloud_storage.addon.id
  }
}

resource "upcloud_server_backup" "example_server_backup" {
  server = upcloud_server.example.id
  plan = "weeklies"
  time = "2200"
}

resource "upcloud_storage_backup" "example_storage_backup" {
  storage = upcloud_storage.addon.id
  time = "1100"
  interval = "mon"
  retention = 2
}
```

However this configuration will work fine, as there is no connection between backed up resources
```hcl
resource "upcloud_storage" "addon" {
  title = "addon"
  size = 10
  zone = "pl-waw1"
}

resource "upcloud_storage_backup" "example_storage_backup" {
  storage = upcloud_storage.addon.id
  time = "1100"
  interval = "mon"
  retention = 2
}

resource "upcloud_server" "example" {
  zone = "pl-waw1"
  plan = "1xCPU-1GB"
  hostname = "mainx1"
  
  template {
    storage = "Ubuntu Server 20.04 LTS (Focal Fossa)"
    size = 10
  }

  network_interface {
    type = "public"
  }
}

resource "upcloud_server_backup" "example_server_backup" {
  server = upcloud_server.example.id
  plan = "weeklies"
  time = "2200"
}
```

