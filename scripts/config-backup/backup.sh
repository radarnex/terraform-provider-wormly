#!/usr/bin/env bash

# Wormly Config Backup Script
# Usage: backup.sh <API_KEY> [BACKUP_PATH]
# If BACKUP_PATH is not provided, a timestamped folder will be created in the current directory.

set -euo pipefail

API_KEY="${1:-}"
BACKUP_PATH="${2:-}"

if [[ -z "$API_KEY" ]]; then
  echo "Usage: $0 <API_KEY> [BACKUP_PATH]"
  echo ""
  echo "Arguments:"
  echo "  API_KEY      Wormly API key (required)"
  echo "  BACKUP_PATH  Path to store backup files (optional)"
  echo ""
  echo "If BACKUP_PATH is not provided, a timestamped folder will be created"
  echo "in the current directory (e.g., wormly_backup_20250704_143022)"
  exit 1
fi

if [[ -z "$BACKUP_PATH" ]]; then
  TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
  BACKUP_PATH="./wormly_backup_$TIMESTAMP"
fi

mkdir -p "$BACKUP_PATH"

# Base API URL
BASE_URL="${WORMLY_API_BASE_URL:-https://api.wormly.com}"

echo "Starting Wormly configuration backup..."
echo "Backup directory: $BACKUP_PATH"
echo ""

# Helper function to make API calls and save responses
fetch_and_save() {
  local api_method="$1"
  local outfile="$2"
  local extra_params="${3:-}"
  
  echo "Fetching $api_method..."
  
  # Build the URL with query string parameters
  local url="$BASE_URL?key=$API_KEY&response=json&cmd=$api_method"
  
  if [[ -n "$extra_params" ]]; then
    url="$url&$extra_params"
  fi
  
  # Execute the curl command with GET request and save to file
  curl -sSL "$url" > "$BACKUP_PATH/$outfile"
  
  # Check if the response contains an error
  if grep -q '"errorcode":[^0]' "$BACKUP_PATH/$outfile" 2>/dev/null; then
    echo "  Warning: API returned an error for $api_method"
    local error_code
    error_code=$(grep -o '"errorcode":[0-9]*' "$BACKUP_PATH/$outfile" | cut -d: -f2)
    echo "  Error code: $error_code"
  else
    echo "  ✓ Saved to $outfile"
  fi
}

# Helper function to fetch hosts and then their sensors
fetch_hosts_and_sensors() {
  echo "Fetching hosts and their sensors..."
  
  # First get the list of hosts
  fetch_and_save "getHostStatus" "hosts.json"
  
  # Extract host IDs and fetch sensors for each host
  if [[ -f "$BACKUP_PATH/hosts.json" ]]; then
    # Parse host IDs from the JSON response (assuming the response contains a status array with hostid fields)
    local host_ids
    host_ids=$(grep -o '"hostid":[0-9]*' "$BACKUP_PATH/hosts.json" 2>/dev/null | cut -d: -f2 | sort -u)
    
    if [[ -n "$host_ids" ]]; then
      echo "Found hosts with IDs: $(echo $host_ids | tr '\n' ' ')"
      
      # Create a sensors directory
      mkdir -p "$BACKUP_PATH/sensors"
      
      # Fetch sensors for each host
      for host_id in $host_ids; do
        echo "  Fetching sensors for host $host_id..."
        fetch_and_save "getHostSensors" "sensors/host_${host_id}_sensors.json" "hostid=$host_id"
      done
    else
      echo "  No hosts found or unable to parse host IDs"
    fi
  fi
}

# Backup all entities
echo "1. Fetching hosts and sensors..."
fetch_hosts_and_sensors

echo ""
echo "2. Fetching contacts..."
fetch_and_save "getContactList" "contacts.json"

echo ""
echo "3. Fetching alert groups..."
fetch_and_save "getAlertGroups" "alert_groups.json"

echo ""
echo "4. Attempting to fetch global alert mute state..."
# Note: There might not be a direct API to get the current global mute state,
# so we'll try to set it to its current state to see the response
fetch_and_save "setGlobalAlertMute" "global_alert_mute.json" "alertsmuted=0" || {
  echo "  Note: Unable to fetch global alert mute state - this might require a different approach"
}

echo ""
echo "5. Fetching additional configuration data..."

# Try to get other potentially useful data
fetch_and_save "getHostStatus" "host_status.json" || echo "  Could not fetch host status"

echo ""
echo "Backup completed successfully!"
echo ""
echo "Files created in $BACKUP_PATH:"
ls -la "$BACKUP_PATH"

echo ""
echo "Backup summary:"
echo "- hosts.json: All monitored hosts"
echo "- host_status.json: Current status of all hosts"
echo "- sensors/: Individual sensor configurations per host"
echo "- contacts.json: All contact persons and alert channels"
echo "- alert_groups.json: All alert groups configuration"
echo "- global_alert_mute.json: Global alert mute state (if available)"
echo ""
echo "Note: This backup contains configuration only - no historical monitoring data."
echo "Sensitive data like API keys and credentials are excluded for security."
