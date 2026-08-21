resource "hcloud_ssh_key" "operator" {
  for_each   = toset(var.admin_ssh_public_keys)
  name       = "${local.name_prefix}-${substr(sha256(each.value), 0, 12)}"
  public_key = each.value
  labels     = local.ownership_labels
}

resource "hcloud_placement_group" "control_plane" {
  name   = "${local.name_prefix}-control-plane"
  type   = "spread"
  labels = local.ownership_labels
}

resource "hcloud_server" "control_plane" {
  count              = var.control_plane_count
  name               = "${local.name_prefix}-control-${format("%02d", count.index + 1)}"
  image              = var.image
  server_type        = var.server_type
  location           = var.location
  ssh_keys           = values(hcloud_ssh_key.operator)[*].id
  firewall_ids       = [hcloud_firewall.nodes.id]
  placement_group_id = hcloud_placement_group.control_plane.id
  labels             = merge(local.ownership_labels, { role = "control-plane" })
  user_data = templatefile("${path.module}/../cloud-init/cloud-config.yaml.tftpl", {
    admin_ssh_public_keys = var.admin_ssh_public_keys
  })

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
}

resource "hcloud_server" "worker" {
  count        = var.worker_count
  name         = "${local.name_prefix}-worker-${format("%02d", count.index + 1)}"
  image        = var.image
  server_type  = var.server_type
  location     = var.location
  ssh_keys     = values(hcloud_ssh_key.operator)[*].id
  firewall_ids = [hcloud_firewall.nodes.id]
  labels       = merge(local.ownership_labels, { role = "worker" })
  user_data = templatefile("${path.module}/../cloud-init/cloud-config.yaml.tftpl", {
    admin_ssh_public_keys = var.admin_ssh_public_keys
  })

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
}

resource "hcloud_server" "bastion" {
  count        = var.enable_temporary_bastion ? 1 : 0
  name         = "${local.name_prefix}-temporary-bastion"
  image        = var.image
  server_type  = var.bastion_server_type
  location     = var.location
  ssh_keys     = values(hcloud_ssh_key.operator)[*].id
  firewall_ids = [hcloud_firewall.bastion[0].id]
  labels       = merge(local.ownership_labels, { role = "bastion", temporary = "true" })
  user_data = templatefile("${path.module}/../cloud-init/cloud-config.yaml.tftpl", {
    admin_ssh_public_keys = var.admin_ssh_public_keys
  })

  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
}
