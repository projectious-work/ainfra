variable "environment_name" {
  description = "Non-secret name retained by the provider-free example."
  type        = string

  validation {
    condition     = length(trimspace(var.environment_name)) > 0
    error_message = "environment_name must not be empty."
  }
}
