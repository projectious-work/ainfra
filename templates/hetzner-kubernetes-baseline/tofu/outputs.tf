output "inventory_nodes" {
  description = "Secret-free node facts used to generate Ansible inventory."
  value = concat(
    [
      for index, server in hcloud_server.control_plane : {
        name         = server.name
        role         = "control-plane-capable"
        private_ipv4 = hcloud_server_network.control_plane[index].ip
        public_ipv4  = var.public_ipv4 ? server.ipv4_address : null
        public_ipv6  = var.public_ipv6 ? server.ipv6_address : null
        image        = var.image
      }
    ],
    [
      for index, server in hcloud_server.worker : {
        name         = server.name
        role         = "worker"
        private_ipv4 = hcloud_server_network.worker[index].ip
        public_ipv4  = var.public_ipv4 ? server.ipv4_address : null
        public_ipv6  = var.public_ipv6 ? server.ipv6_address : null
        image        = var.image
      }
    ],
  )
  sensitive = false
}

output "ownership" {
  description = "Labels that bound lifecycle ownership checks."
  value       = local.ownership_labels
  sensitive   = false
}
