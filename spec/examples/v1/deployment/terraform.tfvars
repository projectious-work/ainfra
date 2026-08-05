environment          = "development"
location             = "fsn1"
control_plane_count  = 1
worker_count         = 0
server_type          = "cx23"
private_cidr         = "10.42.0.0/16"
admin_ssh_public_keys = [
  "ssh-ed25519 PLACEHOLDER operator@example.invalid",
]
