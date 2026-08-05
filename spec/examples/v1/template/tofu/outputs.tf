output "ainfra_inventory" {
  sensitive = false
  value = {
    schema_version = "1"
    groups = {
      control_plane = {
        hosts = {
          example-cp-1 = {
            ansible_host = "10.42.0.10"
            ansible_user = "ainfra"
            vars = {
              ainfra_role = "control-plane"
            }
          }
        }
      }
    }
  }
}
