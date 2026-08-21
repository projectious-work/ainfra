output "ainfra_inventory" {
  description = "Minimal secret-free inventory using private management addresses."
  sensitive   = false
  value = {
    schema_version = "1"
    hosts = merge(
      {
        for index, server in hcloud_server.control_plane : server.name => {
          groups = ["k3s_cluster", "k3s_servers"]
          connection = {
            type    = "ssh"
            address = hcloud_server_network.control_plane[index].ip
            user    = "ainfra"
            port    = 22
          }
        }
      },
      {
        for index, server in hcloud_server.worker : server.name => {
          groups = ["k3s_agents", "k3s_cluster"]
          connection = {
            type    = "ssh"
            address = hcloud_server_network.worker[index].ip
            user    = "ainfra"
            port    = 22
          }
        }
      },
    )
  }
}

output "temporary_bastion" {
  description = "Non-secret temporary bastion facts for initial access and removal review."
  sensitive   = false
  value = var.enable_temporary_bastion ? {
    name        = hcloud_server.bastion[0].name
    public_ipv4 = hcloud_server.bastion[0].ipv4_address
  } : null
}

output "management" {
  description = "Non-secret supported management endpoint after bastion removal."
  sensitive   = false
  value = {
    tunnel_hostname = var.tunnel_hostname
    private_cidr    = var.private_cidr
  }
}

output "ownership" {
  description = "Labels bounding direct provider teardown verification."
  value       = local.ownership_labels
  sensitive   = false
}
