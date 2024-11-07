#!/bin/bash

# Load sensitive keys from JSON file
CONFIG_FILE="keys.json" # Or populate keys_template.json with your keys and rename it to keys.json

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Configuration file $CONFIG_FILE not found."
    exit 1
fi

# Load configuration values
CLOUDFLARE_EMAIL=$(jq -r '.CLOUDFLARE_EMAIL' "$CONFIG_FILE")
CLOUDFLARE_API_KEY=$(jq -r '.CLOUDFLARE_API_KEY' "$CONFIG_FILE")
ZONE_ID=$(jq -r '.ZONE_ID' "$CONFIG_FILE")

# Variables
DOMAIN="${CERTBOT_DOMAIN}"
TOKEN="${CERTBOT_VALIDATION}"
RECORD="_acme-challenge.${DOMAIN}"

# Add TXT record to Cloudflare
RESPONSE=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
    -H "X-Auth-Email: ${CLOUDFLARE_EMAIL}" \
    -H "X-Auth-Key: ${CLOUDFLARE_API_KEY}" \
    -H "Content-Type: application/json" \
    --data "{\"type\":\"TXT\",\"name\":\"${RECORD}\",\"content\":\"${TOKEN}\",\"ttl\":120}")

# Check for errors in the response
if [ "$(echo "$RESPONSE" | jq -r '.success')" != "true" ]; then
    echo "Error adding TXT record to Cloudflare:"
    echo "$RESPONSE" | jq '.errors'
    exit 1
fi

echo "TXT record added successfully."

# Wait for DNS propagation
sleep 25
