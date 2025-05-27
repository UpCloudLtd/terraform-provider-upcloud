variable "zone" {
  type        = string
  description = "UpCloud zone for the supabase deployment"
}

variable "plan" {
  type        = string
  description = "UpCloud plan, e.g. 2xCPU-4GB"
}

variable "ssh_public_key" {
  type        = string
  description = "SSH public key to access the server, e.g. ~/.ssh/id_ed25519.pub"
}

variable "supabase_volume_size" {
  type        = number
  description = "Value in GB for the data volume size, e.g. 50"
}

variable "supabase_template" {
  type        = string
  description = "UpCloud template ID, e.g. 01000000-0000-4000-8000-000030240200 for Ubuntu Server 24.04 LTS (Noble Numbat)"
}

variable "dashboard_username" {
  description = "Username for Supabase Dashboard (only alphanumerics & underscores, 3–16 chars)"
  type        = string

  validation {
    condition = (
      length(var.dashboard_username) >= 3 &&
      length(var.dashboard_username) <= 16 &&
      can(regex("^[_a-zA-Z0-9]+$", var.dashboard_username))
    )
    error_message = "dashboard_username must be 3–16 characters and contain only letters, numbers, or underscores."
  }
}

variable "dashboard_password" {
  description = "Password for Supabase Dashboard (8–32 chars)"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.dashboard_password) >= 8 && length(var.dashboard_password) <= 32
    error_message = "dashboard_password must be between 8 and 32 characters."
  }
}

variable "studio_default_organization" {
  description = "Default Studio organization name (1–64 chars; letters, numbers, spaces, apostrophes, hyphens)"
  type        = string

  validation {
    condition = (
      length(var.studio_default_organization) >= 1 &&
      length(var.studio_default_organization) <= 64 &&
      can(regex("^[A-Za-z0-9' -]+$", var.studio_default_organization))
    )
    error_message = "studio_default_organization must be 1–64 characters and contain only letters, numbers, spaces, apostrophes or hyphens."
  }
}

variable "studio_default_project" {
  description = "Default Studio project name (1–64 chars; letters, numbers, spaces, apostrophes, hyphens)"
  type        = string

  validation {
    condition = (
      length(var.studio_default_project) >= 1 &&
      length(var.studio_default_project) <= 64 &&
      can(regex("^[A-Za-z0-9' -]+$", var.studio_default_project))
    )
    error_message = "studio_default_project must be 1–64 characters and contain only letters, numbers, spaces, apostrophes or hyphens."
  }
}
