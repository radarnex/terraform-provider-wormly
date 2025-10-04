# Data Sources Example

This example demonstrates how to use Wormly data sources to query existing hosts and sensors, and use that information to create new resources.

## What this example does

- Uses `wormly_host` data source to query an existing host
- Uses `wormly_sensor_http` data source to list existing sensors for a host
- Creates a new sensor based on information from the existing host
- Shows how to reference data from existing resources

## Usage

1. Set your Wormly API key:
   ```bash
   export TF_VAR_wormly_api_key="your-api-key-here"
   ```

2. Set the ID of an existing host:
   ```bash
   export TF_VAR_existing_host_id="12345"
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Data Sources Used

- `data.wormly_host.existing` - Queries details about an existing host
- `data.wormly_sensor_http.host_sensors` - Lists all HTTP sensors for a host

## Resources Created

- `wormly_sensor_http.additional_check` - A new sensor created based on existing host data

## Outputs

- `host_name` - Name of the queried host
- `host_enabled` - Whether the host is currently enabled
- `existing_sensors` - List of existing sensors with their details
- `sensor_count` - Total number of existing sensors

## Use Cases

- **Inventory Management**: Query existing monitoring setup
- **Conditional Resources**: Create resources based on existing configuration
- **Integration**: Integrate with existing Wormly setup
- **Reporting**: Generate reports about current monitoring configuration
