# Terraform Provider for Wormly

[![Go Report Card](https://goreportcard.com/badge/github.com/radarnex/terraform-provider-wormly)](https://goreportcard.com/report/github.com/radarnex/terraform-provider-wormly)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

This repository contains a Terraform provider for [Wormly](https://www.wormly.com), a website and server monitoring service. The provider is built using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## Features

- **Resources:**
  - `wormly_host` - Manage monitoring hosts
  - `wormly_sensor_http` - Manage HTTP sensors for hosts
  - `wormly_scheduled_downtime_period` - Manage scheduled maintenance windows for hosts
  - `wormly_global_alerts_mute` - Manage global alert muting settings

- **Data Sources:**
  - `wormly_host` - Query existing host configurations
  - `wormly_sensor_http` - Query existing HTTP sensors

## Roadmap and Status

|  #  | Step                                                      | Status |
| :-: | --------------------------------------------------------- | :----: |
|  1  | Host, Downtime and Global Alerts mute                     |   ✅   |
|  2  | HTTP sensor support                                       |   ✅   |
|  3  | Add more data sources                                     |   ❌   |
|  4  | Do Not Disturb support                                    |   ❌   |
|  5  | FTP and Ping sensor support                               |   ❌   |
|  6  | Health Monitoring support                                 |   ❌   |

## Known issues

The provider relies on Wormly API support, currently there're some known issues. 

- **[Host]** It's not possible to customise any values from the API, so you need to tweak any settings (e.g., `Primary Monitoring Node`, etc) from the UI.
- **[Sensor coverage]** Only HTTP, FTP (pending) and Ping (pending) will be supported in the provider unless the API supports other sensor types.
- **[HTTP sensor updates]** The Wormly API command reference currently exposes `addHostSensor_HTTP`, `getHostSensors`, `enableSensor`, `disableSensor`, and `deleteSensor`, but no dedicated update/edit command for HTTP sensor settings. Because of this, changing HTTP sensor attributes other than `enabled` results in Terraform planning replacement (`delete` + `create`) instead of in-place update.
- **[Host updates]** Wormly API does not currently expose an in-place update command for host name or test interval in this provider integration. Changing `name` or `test_interval` therefore plans replacement; only `enabled` is updated in place.
- **[Scheduled downtime period updates]** Scheduled downtime periods are updatable in place, but changing `hostid` plans replacement.
- **[Global alerts mute drift]** Wormly API does not currently provide a read endpoint for global alert mute state in this provider integration. If the value is changed outside Terraform, drift cannot be detected during refresh.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- A [Wormly](https://www.wormly.com) account and API key

## Installation

### Terraform Registry (Recommended)

This provider will be available on the [Terraform Registry](https://registry.terraform.io/). Add it to your Terraform configuration:

```hcl
terraform {
  required_providers {
    wormly = {
      source = "radarnex/wormly"
      version = "~> 0.1"
    }
  }
}

provider "wormly" {
  api_key = var.wormly_api_key
}
```

### Manual Installation

1. Download the appropriate binary from the [releases page](https://github.com/radarnex/terraform-provider-wormly/releases)
2. Place it in your Terraform plugins directory
3. Run `terraform init`

## Usage

### Quick Start

```hcl
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

# Create a host
resource "wormly_host" "example" {
  name          = "example.com"
  test_interval = 60
  enabled       = true
}

# Create an HTTP sensor
resource "wormly_sensor_http" "example" {
  host_id     = wormly_host.example.id
  nice_name   = "Homepage Check"
  enabled     = true
  url         = "https://example.com"
  method      = "GET"
  timeout     = 30
  check_text  = "Example Domain"
  use_ssl     = true
}

# Schedule daily maintenance downtime
resource "wormly_scheduled_downtime_period" "daily_maintenance" {
  hostid     = wormly_host.example.id
  start      = "02:00"
  end        = "04:00"
  timezone   = "UTC"
  recurrence = "DAILY"
}

# Control global alert muting
resource "wormly_global_alerts_mute" "emergency_mute" {
  enabled = false
}
```

### Provider Configuration

```hcl
provider "wormly" {
  api_key = var.wormly_api_key
  
  # Optional: Custom API endpoint
  base_url = "https://api.wormly.com"
  
  # Optional: Rate limiting (requests per second)
  requests_per_second = 10
  
  # Optional: Retry configuration
  max_retries        = 3
  initial_backoff    = "500ms"
  backoff_multiplier = 1.5
  max_backoff        = "10s"
  
  # Optional: Custom user agent
  user_agent = "terraform-provider-wormly/1.0"
}
```

### Examples

See the [`examples/`](./examples/) directory for complete examples:

- [`basic/`](./examples/basic/) - Basic host and sensor setup
- [`disabled_host/`](./examples/disabled_host/) - Managing disabled hosts
- [`data_sources/`](./examples/data_sources/) - Using data sources
- [`global_alerts_mute/`](./examples/global_alerts_mute/) - Global alert muting configuration
- [`retry_tuning/`](./examples/retry_tuning/) - Custom retry configuration

## Documentation

Complete documentation is available in the [`docs/`](./docs/) directory:

- [Provider Configuration](./docs/index.md)
- [Resources](./docs/resources/)
  - [wormly_host](./docs/resources/host.md)
  - [wormly_sensor_http](./docs/resources/sensor_http.md)
  - [wormly_scheduled_downtime_period](./docs/resources/scheduled_downtime_period.md)
  - [wormly_global_alerts_mute](./docs/resources/global_alerts_mute.md)
- [Data Sources](./docs/data-sources/)
  - [wormly_host](./docs/data-sources/host.md)
  - [wormly_sensor_http](./docs/data-sources/sensor_http.md)

## Contributing

Contributions are welcome! Please see the [Contributing Guide](CONTRIBUTING.md) for development setup, testing, and release process information.

## License

This project is licensed under the MPL 2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: See the [`docs/`](./docs/) directory
- **Examples**: See the [`examples/`](./examples/) directory  
- **Issues**: [GitHub Issues](https://github.com/radarnex/terraform-provider-wormly/issues)
