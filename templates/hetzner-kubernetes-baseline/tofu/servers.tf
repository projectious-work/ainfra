resource "hcloud_ssh_key" "operator" {
  for_each   = toset(var.admin_ssh_public_keys)
  name       = "${local.name_prefix}-${substr(sha256(each.value), 0, 12)}"
  public_key = each.value
  labels     = local.ownership_labels
}

resource "hcloud_server" "control_plane" {
  count       = var.control_plane_count
  name        = "${local.name_prefix}-control-${format("%02d", count.index + 1)}"
  image       = var.image
  server_type = var.server_type
  location    = var.location
  ssh_keys    = values(hcloud_ssh_key.operator)[*].id
  labels      = merge(local.ownership_labels, { role = "control-plane-capable" })
  user_data = templatefile(
    "${path.module}/../cloud-init/cloud-config.yaml.tftpl",
    { admin_ssh_public_keys = var.admin_ssh_public_keys },
  )

  public_net {
    ipv4_enabled = var.public_ipv4
    ipv6_enabled = var.public_ipv6
  }
}

resource "hcloud_server" "worker" {
  count       = var.worker_count
  name        = "${local.name_prefix}-worker-${format("%02d", count.index + 1)}"
  image       = var.image
  server_type = var.server_type
  location    = var.location
  ssh_keys    = values(hcloud_ssh_key.operator)[*].id
  labels      = merge(local.ownership_labels, { role = "worker" })
  user_data = templatefile(
    "${path.module}/../cloud-init/cloud-config.yaml.tftpl",
    { admin_ssh_public_keys = var.admin_ssh_public_keys },
  )

  public_net {
    ipv4_enabled = var.public_ipv4
    ipv6_enabled = var.public_ipv6
  }
}
