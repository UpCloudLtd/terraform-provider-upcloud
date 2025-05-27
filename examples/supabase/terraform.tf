terraform {
  required_providers {
    upcloud = {
      source  = "registry.upcloud.com/upcloud/upcloud"
      version = "~> 5.22.0"
    }
    cloudinit = {
      source  = "hashicorp/cloudinit"
      version = "~> 2.3"
    }
  }
}

locals {
  stripped_disk_id  = replace(upcloud_storage.supabase_data_volume.id, "-", "")
  truncated_disk_id = substr(local.stripped_disk_id, 0, 20) # UpCloud truncates to 20 chars
  disk_path         = "/dev/disk/by-id/virtio-${local.truncated_disk_id}"
}

# Data source for Ubuntu cloud image (UpCloud template)
data "upcloud_storage" "ubuntu_image" {
  type = "template"
  id   = var.supabase_template
}

# Create UpCloud storage volume for database persistence
resource "upcloud_storage" "supabase_data_volume" {
  size  = var.supabase_volume_size
  zone  = var.zone
  tier  = "standard"
  title = "supabase-data-volume"
}

# Create the UpCloud server for Supabase
resource "upcloud_server" "supabase_server" {
  zone     = var.zone
  plan     = var.plan
  hostname = "supabase-node"
  metadata = true

  network_interface {
    type = "public"
  }

  template {
    storage = data.upcloud_storage.ubuntu_image.id
    size    = 25
  }

  # Attach the data volume as secondary disk
  storage_devices {
    storage          = upcloud_storage.supabase_data_volume.id
    type             = "disk"
    address          = "virtio"
    address_position = 1
  }

  login {
    user            = "root"
    keys            = [file(var.ssh_public_key)]
    create_password = false
  }

  user_data = data.template_cloudinit_config.supabase_cloudinit.rendered

}

# Cloud-init script to set up Supabase
data "template_cloudinit_config" "supabase_cloudinit" {
  gzip          = false
  base64_encode = false

  part {
    filename     = "setup.sh"
    content_type = "text/x-shellscript"
    content      = <<-EOT
      #!/bin/bash
      set -x
      echo "[cloud-init] Installing Docker..." 
      apt-get update
      apt-get install -y ca-certificates curl gnupg lsb-release

      install -m 0755 -d /etc/apt/keyrings
      curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
      chmod a+r /etc/apt/keyrings/docker.asc

      arch="$(dpkg --print-architecture)"
      codename="$(. /etc/os-release && echo "$${UBUNTU_CODENAME:-$${VERSION_CODENAME}}")"

      echo "deb [arch=$${arch} signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $${codename} stable" \
      > /etc/apt/sources.list.d/docker.list

      apt-get update -qq
      apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin git

      echo "[cloud-init] Installing Git..."
      apt-get install -y -qq git-all

      systemctl enable docker
      systemctl start docker

      echo "[cloud-init] Mounting attached storage volume for Postgres data..."
      mkfs.ext4 ${local.disk_path} || true
      mkdir -p /supabase
      mount ${local.disk_path} /supabase      

      echo "[cloud-init] Get Supabase official repo..."
      cd /supabase
      git clone --depth 1 https://github.com/supabase/supabase

      echo "[cloud-init] Create supabase project directory..."
      mkdir supabase-project

      echo "[cloud-init] Copying docker-compose.yml to project directory..."
      cp -rf supabase/docker/* supabase-project

      echo "[cloud-init] Copy the fake env vars..."
      cp supabase/docker/.env.example supabase-project/.env


      echo "[cloud-init] Templating .env with provided Terraform variables..."
      sed -i "s|^DASHBOARD_USERNAME=.*|DASHBOARD_USERNAME=${var.dashboard_username}|" supabase-project/.env
      sed -i "s|^DASHBOARD_PASSWORD=.*|DASHBOARD_PASSWORD=${var.dashboard_password}|" supabase-project/.env
      sed -i "s|^STUDIO_DEFAULT_ORGANIZATION=.*|STUDIO_DEFAULT_ORGANIZATION=\"${var.studio_default_organization}\"|" supabase-project/.env
      sed -i "s|^STUDIO_DEFAULT_PROJECT=.*|STUDIO_DEFAULT_PROJECT=\"${var.studio_default_project}\"|"    supabase-project/.env


      echo "[cloud-init] Pull the latest Docker images..."
      cd supabase-project
      docker compose pull

      echo [cloud-init] "start the services in detached mode..."
      docker compose up -d

    EOT
  }
}


