# upcloud_storage_backup (Resource)

The UpCloud `storage_backup resource allows managing backup rules for individual storages

## Example

The following example shows the creation of a `storage` resource and a `storage_backup` resource to set the proper backup rules. With this configuration, the storage backups will be taken each Monday at 22pm and will be kept for 7 days.

```hcl
resource "upcloud_storage" "example_storage" {
  size  = 10
  tier  = "maxiops"
  title = "My data collection"
  zone  = "fi-hel1"
}

resource "upcloud_storage_backup" "example_backup" {
  storage = upcloud_storage.example_storage.id
  time = "2200"
  interval = "mon"
  retention = 7
}
```

## Schema

### Required

- **storage** (String) The ID of the storage that should be backed up
- **time** (String) Exact hour at which the backup should be taken in hhmm format
- **interval** (String) The day of the week on which backup will be taken. Can also specify it as "daily" to take backups everyday.
- **retention** (Int) The amount of days the backup should be kept

## Notes

### Conflicting backup rules
Please note that it is impossible to use `storage_backup` and `server_backup` resources for the same resources. Trying to set `storage_backup` for a storage that is attached to a server that already has a `server_backup` will fail. For examples, please see [server_backup resource docs](./storage_backup.md)
