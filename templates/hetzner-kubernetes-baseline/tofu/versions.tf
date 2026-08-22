terraform {
  required_version = ">= 1.10.0, < 2.0.0"

  backend "local" {}

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.64.0"
    }
  }
}
