resource "terraform_data" "clean_room" {
  input = {
    environment_name = var.environment_name
  }
}
