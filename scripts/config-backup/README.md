# Wormly Configuration Backup Script

This script creates a backup of your Wormly account configuration, including hosts, sensors, contacts, alert groups, and other settings.

## Features

- ✅ Backs up all monitored hosts
- ✅ Backs up all sensors (per host)
- ✅ Backs up all contacts and alert channels
- ✅ Backs up all alert groups
- ✅ Backs up global alert mute state (if available)
- ✅ Configuration only (no historical monitoring data)
- ✅ JSON format for easy parsing and analysis
- ✅ Timestamped backup folders
- ✅ Excludes sensitive data for security

## Usage

```bash
./backup.sh <API_KEY> [BACKUP_PATH]
```

### Arguments

- `API_KEY` (required): Your Wormly API key
- `BACKUP_PATH` (optional): Directory where backup files will be stored

### Examples

```bash
# Basic usage - creates timestamped folder in current directory
./backup.sh your_api_key_here

# Specify custom backup directory
./backup.sh your_api_key_here /path/to/backup/directory

# Using environment variable for API key
./backup.sh "$WORMLY_API_KEY" ./backups/$(date +%Y%m%d)
```

## Output Structure

The script creates the following backup structure:

```
wormly_backup_20250704_143022/
├── hosts.json                      # All monitored hosts
├── host_status.json                # Current status of all hosts
├── contacts.json                   # All contact persons and alert channels
├── alert_groups.json               # All alert groups configuration
├── global_alert_mute.json          # Global alert mute state
└── sensors/                        # Per-host sensor configurations
    ├── host_123_sensors.json
    ├── host_456_sensors.json
    └── ...
```

## File Descriptions

- **hosts.json**: Contains all monitored hosts with their basic configuration
- **host_status.json**: Current monitoring status of all hosts
- **sensors/**: Individual JSON files containing sensor configurations for each host
- **contacts.json**: All contact persons and their configured alert channels
- **alert_groups.json**: Alert group configurations and escalation settings
- **global_alert_mute.json**: Global alert mute state (if API supports fetching it)

## API Endpoints Used

The script uses the following Wormly API endpoints:

- `getHostStatus` - Fetch host information and status
- `getHostSensors` - Fetch sensors for each host
- `getContactList` - Fetch contact persons and alert channels
- `getAlertGroups` - Fetch alert groups configuration
- `setGlobalAlertMute` - Attempt to fetch global alert mute state

## Security Notes

- The script **excludes sensitive data** like API keys and credentials from the backup
- Store your API key securely and never commit it to version control
- Consider encrypting backup files if they contain sensitive configuration data
- Backup files contain configuration that could be used to understand your monitoring setup

## Error Handling

- The script will continue if individual API calls fail
- Warnings are displayed for any API errors encountered
- Error codes from the Wormly API are reported when available
- The script uses `set -euo pipefail` for robust error handling

## Troubleshooting

### Common Issues

1. **"API returned an error"**
   - Check that your API key is valid and has the necessary permissions
   - Verify you're not hitting API rate limits

2. **"No hosts found"**
   - Ensure you have hosts configured in your Wormly account
   - Check that the API key has permission to read host information

3. **"Could not fetch ..."**
   - Some API endpoints might not be available depending on your account type
   - Check the Wormly API documentation for endpoint availability

### Getting API Key

1. Log into your Wormly account
2. Go to [API Keys](https://www.wormly.com/apikeys)
3. Generate a new API key or use an existing one

## Related

- [Wormly API Reference](https://www.wormly.com/api_reference)
- [Wormly Developers Guide](https://www.wormly.com/developers)
- [Terraform Provider for Wormly](../../README.md)
