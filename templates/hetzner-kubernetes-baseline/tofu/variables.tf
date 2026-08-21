variable "environment" {
  description = "Lowercase deployment name used in resource names and ownership labels."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,30}$", var.environment))
    error_message = "environment must be a lowercase DNS-style name of 2-31 characters."
  }
}

variable "location" {
  description = "Hetzner location for all servers."
  type        = string
  default     = "nbg1"

  validation {
    condition     = contains(["fsn1", "nbg1", "hel1"], var.location)
    error_message = "location must be fsn1, nbg1, or hel1."
  }
}

variable "control_plane_count" {
  description = "Odd number of K3s server nodes."
  type        = number
  default     = 3

  validation {
    condition     = contains([1, 3, 5], var.control_plane_count)
    error_message = "control_plane_count must be 1, 3, or 5."
  }
}

variable "worker_count" {
  description = "Optional K3s agent node count."
  type        = number
  default     = 0

  validation {
    condition     = var.worker_count >= 0 && var.worker_count <= 20 && floor(var.worker_count) == var.worker_count
    error_message = "worker_count must be an integer from 0 through 20."
  }
}

variable "image" {
  description = "Pinned supported Hetzner image alias."
  type        = string
  default     = "debian-13"

  validation {
    condition     = var.image == "debian-13"
    error_message = "the only supported image is debian-13."
  }
}

variable "server_type" {
  description = "Hetzner server type for cluster nodes."
  type        = string
  default     = "cx23"
}

variable "bastion_server_type" {
  description = "Hetzner server type for the temporary bastion."
  type        = string
  default     = "cx23"
}

variable "private_cidr" {
  description = "RFC1918 private network, /16 through /24."
  type        = string
  default     = "10.42.0.0/16"

  validation {
    condition = can(cidrnetmask(var.private_cidr)) && can(
      regex("^(10\\.|192\\.168\\.|172\\.(1[6-9]|2[0-9]|3[01])\\.)", var.private_cidr)
      ) && tonumber(split("/", var.private_cidr)[1]) >= 16 && (
      tonumber(split("/", var.private_cidr)[1]) <= 24
    )
    error_message = "private_cidr must be RFC1918 IPv4 between /16 and /24."
  }
}

variable "admin_ssh_public_keys" {
  description = "Operator-supplied Ed25519 public keys; private keys are never generated."
  type        = list(string)

  validation {
    condition = length(var.admin_ssh_public_keys) > 0 && alltrue([
      for key in var.admin_ssh_public_keys :
      startswith(key, "ssh-ed25519 ") || startswith(key, "sk-ssh-ed25519@openssh.com ")
    ])
    error_message = "at least one Ed25519 public key is required."
  }
}

variable "enable_temporary_bastion" {
  description = "Create the temporary public SSH bastion. Disable it after tunnel verification."
  type        = bool
  default     = false
}

variable "bastion_admin_cidrs" {
  description = "Narrow source CIDRs permitted to reach the temporary bastion on TCP/22."
  type        = list(string)
  default     = []

  validation {
    condition = alltrue([
      for cidr in var.bastion_admin_cidrs : can(cidrhost(cidr, 0)) &&
      tonumber(split("/", cidr)[1]) >= (strcontains(cidr, ":") ? 64 : 24)
    ])
    error_message = "bastion_admin_cidrs must be IPv4 /24 or IPv6 /64 or narrower."
  }

  validation {
    condition     = !var.enable_temporary_bastion || length(var.bastion_admin_cidrs) > 0
    error_message = "enable_temporary_bastion requires at least one narrow admin CIDR."
  }
}

variable "tunnel_hostname" {
  description = "Externally managed Cloudflare Tunnel hostname used after bastion removal."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?$", var.tunnel_hostname))
    error_message = "tunnel_hostname must be a DNS hostname."
  }
}
