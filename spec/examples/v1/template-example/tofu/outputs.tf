output "ainfra_inventory" {
  description = "Provider-free localhost inventory for the conformance lifecycle."
  sensitive   = false
  value = {
    schema_version = "1"
    groups = {
      example = {
        hosts = {
          (terraform_data.example.output.host_name) = {
            ansible_host       = "127.0.0.1"
            ansible_connection = "local"
            vars = {
              ainfra_environment = terraform_data.example.output.environment
            }
          }
        }
      }
    }
  }
}
