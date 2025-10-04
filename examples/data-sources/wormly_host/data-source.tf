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

# Query an existing host
data "wormly_host" "existing" {
  id = var.existing_host_id
}

# Output host information
output "host_name" {
  description = "Name of the queried host"
  value       = data.wormly_host.existing.name
}

output "host_enabled" {
  description = "Whether the host is enabled"
  value       = data.wormly_host.existing.enabled
}
