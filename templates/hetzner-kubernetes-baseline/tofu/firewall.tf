resource "hcloud_firewall" "nodes" {
  name   = "${local.name_prefix}-nodes"
  labels = local.ownership_labels

  dynamic "rule" {
    for_each = length(var.management_ingress_cidrs) > 0 && (
      var.public_ipv4 || var.public_ipv6
    ) ? [1] : []
    content {
      direction   = "in"
      protocol    = "tcp"
      port        = "22"
      source_ips  = var.management_ingress_cidrs
      description = "Explicit operator management ingress"
    }
  }

  dynamic "rule" {
    for_each = length(var.workload_ingress_cidrs) > 0 && (
      var.public_ipv4 || var.public_ipv6
    ) ? [1] : []
    content {
      direction   = "in"
      protocol    = "tcp"
      port        = "80"
      source_ips  = var.workload_ingress_cidrs
      description = "Explicit workload HTTP ingress"
    }
  }

  dynamic "rule" {
    for_each = length(var.workload_ingress_cidrs) > 0 && (
      var.public_ipv4 || var.public_ipv6
    ) ? [1] : []
    content {
      direction   = "in"
      protocol    = "tcp"
      port        = "443"
      source_ips  = var.workload_ingress_cidrs
      description = "Explicit workload HTTPS ingress"
    }
  }
}
