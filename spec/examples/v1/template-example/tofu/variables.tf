variable "environment" {
  description = "Environment label exposed to the localhost Ansible run."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,31}$", var.environment))
    error_message = "environment must be 2-32 lowercase alphanumeric or hyphen characters and start with a letter."
  }
}

variable "host_name" {
  description = "Ansible inventory name for the local conformance host."
  type        = string
  default     = "localhost"

  validation {
    condition     = can(regex("^[A-Za-z0-9_][A-Za-z0-9_.-]{0,62}$", var.host_name))
    error_message = "host_name must be a valid, non-empty Ansible inventory host name of at most 63 characters."
  }
}
