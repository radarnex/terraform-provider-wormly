terraform {
  required_providers {
    wormly = {
      source = "radarnex/wormly"
    }
  }
}
provider "wormly" {
  # Rate limiting: 5 requests per second (conservative)
  requests_per_second = 5

  # Custom user agent
  user_agent = "terraform-provider-wormly/1.0 (custom-config)"
}
