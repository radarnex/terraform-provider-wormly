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

# Create an HTTP sensor for the host
resource "wormly_sensor_http" "example" {
  host_id         = wormly_host.example.id
  nice_name       = "Homepage Check 2"
  url             = "https://example.com"
  timeout         = 30
  expected_text   = "Example Domain"
  verify_ssl_cert = true
}

# Output the created resources
output "host_id" {
  description = "ID of the created host"
  value       = wormly_host.example.id
}

output "sensor_id" {
  description = "ID of the created HTTP sensor"
  value       = wormly_sensor_http.example.id
}
