locals {
  name_prefix = "ainfra-${var.environment}"
  ownership_labels = {
    "managed-by"  = "ainfra"
    "template"    = "hetzner-kubernetes-baseline"
    "environment" = var.environment
  }
  all_server_ids = concat(
    hcloud_server.control_plane[*].id,
    hcloud_server.worker[*].id,
  )
}
