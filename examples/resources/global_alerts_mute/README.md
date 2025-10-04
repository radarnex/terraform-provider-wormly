# Global Alerts Mute Example

This example demonstrates how to use the `wormly_global_alerts_mute` resource to control the global alert mute setting in Wormly.

## Configuration

The resource has a single attribute:

- `enabled` (bool, optional) - Whether global alerts mute is enabled. Defaults to `false`.

## Usage

```hcl
resource "wormly_global_alerts_mute" "mute" {
  enabled = true
}
```

## Important Notes

- This is a singleton resource - there can only be one instance per provider configuration
- The resource has a fixed ID of "global_alerts_mute"
- When the resource is destroyed, global alerts mute is automatically disabled
- There is no data source for this resource since the Wormly API doesn't provide a way to read the current state
- Import is not supported for this resource since the state cannot be queried from the API

## Running the Example

1. Set your Wormly API key:
   ```bash
   export TF_VAR_wormly_api_key="your-api-key-here"
   ```

2. Initialize and apply:
   ```bash
   terraform init
   terraform apply
   ```

3. To disable global alerts mute, either set `enabled = false` or destroy the resource:
   ```bash
   terraform destroy
   ```
