# Supabase Self-Hosted on UpCloud

This repository provides Terraform configuration to deploy a self-hosted Supabase instance on UpCloud.

## Overview

Using Terraform and cloud-init, this setup will:

* Create an UpCloud storage volume for Postgres data persistence.
* Provision an Ubuntu 24.04 LTS server on UpCloud.
* Format and mount the attached volume at `/supabase` for the database.
* Install Docker, Docker Compose, and Git via cloud-init.
* Clone the official Supabase repository and copy Docker Compose files.
* Generate a `.env` file with your provided credentials for:

  * **Dashboard basic auth** (username and password).
  * **Studio** default organization and project names.
* Pull and start all Supabase services in detached mode using Docker Compose.

## Prerequisites

* **Terraform** v1.x installed locally.
* **UpCloud account** with API credentials configured (see [UpCloud Terraform Provider](https://registry.upcloud.com/upcloud/upcloud/latest/docs)).
* **SSH key pair** (public key path specified in `ssh_public_key`).

## Configuring Input Variables

In terraform.tfvars you can set up important properties of your deployment

| Variable                      | Description                                                                                |
| ----------------------------- | ------------------------------------------------------------------------------------------ |
| `zone`                        | UpCloud zone (e.g., `fi-hel1`).                                                            |
| `plan`                        | UpCloud plan (e.g., `2xCPU-4GB`).                                                          |
| `ssh_public_key`              | Path to your SSH public key (e.g., `~/.ssh/id_ed25519.pub`).                               |
| `supabase_volume_size`        | Size (GB) of the data volume for Postgres (e.g., `50`).                                    |
| `supabase_template`           | UUID of the Ubuntu template (e.g., for 24.04 LTS: `01000000-0000-4000-8000-000030240200`). |
| `dashboard_username`          | Username for Supabase Studio basic auth.                                                   |
| `dashboard_password`          | Password for Supabase Studio basic auth.                                                   |
| `studio_default_organization` | Default organization name in Studio (supports spaces/apostrophes).                         |
| `studio_default_project`      | Default project name in Studio (supports spaces/apostrophes).                              |

Example `terraform.tfvars`:

```hcl
zone                         = "fi-hel1"
plan                         = "2xCPU-4GB"
ssh_public_key               = "~/.ssh/id_ed25519.pub"
supabase_volume_size         = 50
supabase_template            = "01000000-0000-4000-8000-000030240200"
dashboard_username           = "supabase_admin"
dashboard_password           = "SupabaseAdmin123!"
studio_default_organization  = "My Supabase Org"
studio_default_project       = "My Supabase Project"
```

Export your UpCloud credentials:

```bash
export UPCLOUD_USERNAME="<your-upcloud-username>"
export UPCLOUD_PASSWORD="<your-upcloud-password>"
```

## Dashboard Authentication (Production Hardening)

Supabase Studio (the Dashboard) uses HTTP basic auth by default and you can set up login in the terraform.tfvars file.
There are some aspects you have to take into account when going to production and they are details here: ([supabase.com](https://supabase.com/docs/guides/self-hosting/docker))

## Terraform Commands

* **Initialize**:

  ```bash
  terraform init
  ```

* **Plan**:

  ```bash
  terraform plan -var-file="terraform.tfvars"
  ```

* **Apply**:

  ```bash
  terraform apply -var-file="terraform.tfvars"
  ```

* **Destroy**:

  ```bash
  terraform destroy -var-file="terraform.tfvars"
  ```

Follow these steps to deploy your self-hosted Supabase instance on UpCloud. Enjoy!
