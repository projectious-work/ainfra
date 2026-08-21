resource "hcloud_network" "private" {
  name     = "${local.name_prefix}-private"
  ip_range = var.private_cidr
  labels   = local.ownership_labels
}

resource "hcloud_network_subnet" "nodes" {
  network_id   = hcloud_network.private.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = var.private_cidr
}

resource "hcloud_server_network" "control_plane" {
  count      = var.control_plane_count
  server_id  = hcloud_server.control_plane[count.index].id
  network_id = hcloud_network.private.id

  depends_on = [hcloud_network_subnet.nodes]
}

resource "hcloud_server_network" "worker" {
  count      = var.worker_count
  server_id  = hcloud_server.worker[count.index].id
  network_id = hcloud_network.private.id

  depends_on = [hcloud_network_subnet.nodes]
}

resource "hcloud_server_network" "bastion" {
  count      = var.enable_temporary_bastion ? 1 : 0
  server_id  = hcloud_server.bastion[0].id
  network_id = hcloud_network.private.id

  depends_on = [hcloud_network_subnet.nodes]
}
