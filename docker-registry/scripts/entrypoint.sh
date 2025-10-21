#!/bin/sh
set -e

# Set default values
export EXTERNAL_URL=${EXTERNAL_URL:-"https://localhost"}
export EXTERNAL_REGISTRY_URL=${EXTERNAL_REGISTRY_URL:-"$EXTERNAL_URL"}
export TRUST_FORWARDED_HEADERS=${TRUST_FORWARDED_HEADERS:-"false"}
export DISABLE_INTERNAL_SSL=${DISABLE_INTERNAL_SSL:-"false"}
export LISTEN_PORT=${LISTEN_PORT:-"8080"}
export JWT_SECRET=${JWT_SECRET:-"$(openssl rand -base64 32)"}

echo "Configuration:"
echo "  External URL: $EXTERNAL_URL"
echo "  Registry URL: $EXTERNAL_REGISTRY_URL"
echo "  Trust Forwarded Headers: $TRUST_FORWARDED_HEADERS"
echo "  Disable Internal SSL: $DISABLE_INTERNAL_SSL"
echo "  Listen Port: $LISTEN_PORT"

# Initialize database if it doesn't exist
if [ ! -f "/database/registry.db" ]; then
    echo "Creating empty database file..."
    touch /database/registry.db
    echo "Database tables will be created by the auth service on first startup"
fi

# Generate configuration files from templates
echo "Generating configuration files..."

# Registry configuration
envsubst < /etc/docker/registry/config.yml.template > /etc/docker/registry/config.yml

# Nginx configuration is now static in /etc/nginx/http.d/default.conf
echo "Using static nginx configuration"

# Ensure proper permissions
chown -R root:root /database /certs /var/lib/registry
chmod 600 /certs/registry.key 2>/dev/null || true

# Start services
echo "Starting services..."
exec "$@"