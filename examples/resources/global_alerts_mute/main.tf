# Example usage of the wormly_global_alerts_mute resource

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

# Enable global alerts mute
resource "wormly_global_alerts_mute" "mute" {
  enabled = true
}

# Variables
variable "wormly_api_key" {
  description = "Wormly API key"
  type        = string
  sensitive   = true
}
