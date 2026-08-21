environment = "replace-phase9"
location    = "nbg1"

control_plane_count = 3
worker_count        = 0
server_type         = "cx23"

admin_ssh_public_keys = [
  "ssh-ed25519 REPLACE_WITH_OPERATOR_PUBLIC_KEY phase9-example",
]

private_cidr             = "10.42.0.0/16"
tunnel_hostname          = "k3s-admin.example.invalid"
enable_temporary_bastion = false
bastion_admin_cidrs      = []
