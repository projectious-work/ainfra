variable "environment" {
  description = "Environment name used in resource identity and ownership."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,62}$", var.environment))
    error_message = "environment must be a lowercase DNS-style name."
  }
}

variable "location" {
  description = "Hetzner location: fsn1, nbg1, or hel1."
  type        = string

  validation {
    condition     = contains(["fsn1", "nbg1", "hel1"], var.location)
    error_message = "location must be fsn1, nbg1, or hel1."
  }
}

variable "control_plane_count" {
  description = "Control-plane-capable hosts; v0.1 supports exactly one."
  type        = number
  default     = 1

  validation {
    condition     = var.control_plane_count == 1
    error_message = "v0.1 requires exactly one control-plane-capable host."
  }
}

variable "worker_count" {
  description = "Optional worker host count."
  type        = number
  default     = 0

  validation {
    condition     = var.worker_count >= 0 && floor(var.worker_count) == var.worker_count
    error_message = "worker_count must be a non-negative integer."
  }
}

variable "image" {
  description = "Supported images: debian-13 (the only v0.1 option)."
  type        = string
  default     = "debian-13"

  validation {
    condition     = var.image == "debian-13"
    error_message = "the only supported image is debian-13."
  }
}

variable "server_type" {
  description = "Hetzner server type selected by the operator."
  type        = string
  default     = "cx23"
}

variable "private_cidr" {
  description = "Narrow RFC1918 private network, /16 or smaller."
  type        = string
  default     = "10.42.0.0/16"

  validation {
    condition = can(cidrnetmask(var.private_cidr)) && (
      tonumber(split("/", var.private_cidr)[1]) >= 16
    )
    error_message = "private_cidr must be IPv4 /16 or narrower."
  }
}

variable "public_ipv4" {
  description = "Allocate a billable public IPv4 address when true."
  type        = bool
  default     = false
}

variable "public_ipv6" {
  description = "Allocate public IPv6 connectivity when true."
  type        = bool
  default     = false
}

variable "management_ingress_cidrs" {
  description = "Explicit narrow CIDRs allowed to reach public SSH."
  type        = list(string)
  default     = []
}

variable "workload_ingress_cidrs" {
  description = "Explicit CIDRs allowed to reach public HTTP/HTTPS."
  type        = list(string)
  default     = []
}

variable "admin_ssh_public_keys" {
  description = "Operator-supplied public keys; private keys are prohibited."
  type        = list(string)
  sensitive   = false

  validation {
    condition = length(var.admin_ssh_public_keys) > 0 && alltrue([
      for key in var.admin_ssh_public_keys :
      startswith(key, "ssh-ed25519 ") ||
      startswith(key, "sk-ssh-ed25519@openssh.com ")
    ])
    error_message = "at least one supported Ed25519 public key is required."
  }
}
