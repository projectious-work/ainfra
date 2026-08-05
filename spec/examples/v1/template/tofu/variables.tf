variable "environment" {
  type = string
}

variable "location" {
  type = string
}

variable "control_plane_count" {
  type = number
}

variable "worker_count" {
  type = number
}

variable "server_type" {
  type = string
}

variable "private_cidr" {
  type = string
}

variable "admin_ssh_public_keys" {
  type      = list(string)
  sensitive = false
}
