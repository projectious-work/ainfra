locals {
  name_prefix = "ainfra-${var.environment}"
  ownership_labels = {
    "managed-by"  = "ainfra"
    "template"    = "hetzner-kubernetes-baseline"
    "environment" = var.environment
  }
}
