terraform {
  required_providers {
    wormly = {
      source = "radarnex/wormly"
    }
  }
}

provider "wormly" {
  api_key = var.wormly_api_key
}

variable "wormly_api_key" {
  description = "Wormly API key"
  type        = string
  sensitive   = true
}

variable "existing_host_id" {
  description = "ID of an existing Wormly host to query"
  type        = string
}

# Query sensors for the host
data "wormly_sensor_http" "host_sensors" {
  host_id = var.existing_host_id
}

# Create a new sensor based on existing host data
resource "wormly_sensor_http" "additional_check" {
  host_id       = var.existing_host_id
  nice_name     = "Additional Check"
  enabled       = "true"
  url           = "https://example.com/health"
  timeout       = 30
  expected_text = "Example"
}

output "existing_sensors" {
  description = "List of existing sensors for the host"
  value = [
    for sensor in data.wormly_sensor_http.host_sensors.sensors : {
      id        = sensor.id
      nice_name = sensor.nice_name
      enabled   = sensor.enabled
    }
  ]
}

output "sensor_count" {
  description = "Number of existing sensors"
  value       = length(data.wormly_sensor_http.host_sensors.sensors)
}
