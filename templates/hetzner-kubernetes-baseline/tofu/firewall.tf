resource "hcloud_firewall" "nodes" {
  name   = "${local.name_prefix}-nodes"
  labels = local.ownership_labels

  rule {
    direction   = "in"
    protocol    = "icmp"
    source_ips  = [var.private_cidr]
    description = "Private-network diagnostics only"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "22"
    source_ips  = [var.private_cidr]
    description = "Private SSH management"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "2379-2380"
    source_ips  = [var.private_cidr]
    description = "Private K3s etcd peer and client traffic"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "6443"
    source_ips  = [var.private_cidr]
    description = "Private Kubernetes API traffic"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "10250"
    source_ips  = [var.private_cidr]
    description = "Private kubelet traffic"
  }

  rule {
    direction   = "in"
    protocol    = "udp"
    port        = "8472"
    source_ips  = [var.private_cidr]
    description = "Private Flannel VXLAN traffic"
  }
}

resource "hcloud_firewall" "bastion" {
  count  = var.enable_temporary_bastion ? 1 : 0
  name   = "${local.name_prefix}-temporary-bastion"
  labels = merge(local.ownership_labels, { temporary = "true" })

  dynamic "rule" {
    for_each = var.bastion_admin_cidrs
    content {
      direction   = "in"
      protocol    = "tcp"
      port        = "22"
      source_ips  = [rule.value]
      description = "Temporary operator SSH ingress"
    }
  }
}
