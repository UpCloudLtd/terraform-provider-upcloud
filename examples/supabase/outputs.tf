# Output the useful information
output "supabase_ip" {
  value = upcloud_server.supabase_server.network_interface[0].ip_address
}

output "volume_id" {
  value = upcloud_storage.supabase_data_volume.id
}

output "out_disk_id" {
  value = local.disk_path
}