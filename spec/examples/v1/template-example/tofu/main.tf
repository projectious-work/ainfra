resource "terraform_data" "example" {
  input = {
    environment = var.environment
    host_name   = var.host_name
  }
}
