#TO RESET SSL CERTIFICATES

# Variables
DOMAIN="${CERTBOT_DOMAIN}"
TOKEN="${CERTBOT_VALIDATION}"
RECORD="_acme-challenge.${DOMAIN}"

# Cloudflare API credentials
CLOUDFLARE_EMAIL="kobenaidun@gmail.com"
CLOUDFLARE_API_KEY="083c5ca93134656f451044d76e44ea2a4172c"
ZONE_ID="88f7983cb0dc455e60bf1ca2fd43161f"

# Add TXT record to Cloudflare
curl -s -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
    -H "X-Auth-Email: ${CLOUDFLARE_EMAIL}" \
    -H "X-Auth-Key: ${CLOUDFLARE_API_KEY}" \
    -H "Content-Type: application/json" \
    --data "{\"type\":\"TXT\",\"name\":\"${RECORD}\",\"content\":\"${TOKEN}\",\"ttl\":120}"

# Wait for DNS propagation
sleep 25