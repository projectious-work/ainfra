output "ainfra_inventory" {
  description = "Provider-free localhost inventory for the conformance lifecycle."
  sensitive   = false
  value = {
    schema_version = "1"
    hosts = {
      (terraform_data.example.output.host_name) = {
        groups = ["example"]
        connection = {
          type = "local"
        }
      }
    }
  }
}
