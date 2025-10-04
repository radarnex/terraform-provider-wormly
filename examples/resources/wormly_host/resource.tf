terraform {
  required_providers {
    wormly = {
      source = "radarnex/wormly"
    }
  }
}

provider "wormly" {
  api_key = var.wormly_api_key
  debug   = true
}

variable "wormly_api_key" {
  description = "Wormly API key"
  type        = string
  sensitive   = true
}

# Create a host
resource "wormly_host" "example" {
  name = "example"
  # enabled       = false
  test_interval = 60
}

# Output the created resources
output "host_id" {
  description = "ID of the created host"
  value       = wormly_host.example.id
}
