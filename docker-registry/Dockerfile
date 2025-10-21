# Multi-stage build for optimal image size and security

# Stage 1: Build Go auth service
FROM golang:1.22-bullseye AS auth-builder
WORKDIR /app
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

COPY auth-service/ .
RUN go mod download && \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o auth-service .

# Stage 2: Build Web UI
FROM node:18-alpine AS ui-builder
WORKDIR /app
COPY web-ui/package*.json ./
RUN npm ci --only=production || echo "No package.json found, skipping npm install"
COPY web-ui/ .
RUN mkdir -p dist && cp -r static dist/ && cp index.html dist/

# Stage 3: Final runtime image
FROM registry:2.8.3

# Install required packages
RUN apk add --no-cache \
    nginx \
    supervisor \
    sqlite \
    openssl \
    curl \
    ca-certificates \
    tzdata \
    gettext

# Create required directories
RUN mkdir -p \
    /etc/docker/registry \
    /etc/nginx/http.d \
    /var/log/supervisor \
    /auth \
    /certs \
    /database \
    /var/www/html \
    /scripts

# Copy built binaries and assets
COPY --from=auth-builder /app/auth-service /auth/
COPY --from=ui-builder /app/dist/ /var/www/html/
COPY config/registry.yml.template /etc/docker/registry/config.yml.template
COPY nginx/simple.conf /etc/nginx/http.d/default.conf
COPY scripts/ /scripts/

# Generate SSL certificates
RUN openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /certs/registry.key \
    -out /certs/registry.crt \
    -subj "/C=US/ST=State/L=City/O=Registry/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:registry,IP:127.0.0.1"

# Set permissions
RUN chmod +x /auth/auth-service /scripts/entrypoint.sh && \
    chown -R root:root /certs /auth /database

# Create supervisor configuration
RUN cat > /etc/supervisord.conf <<EOF
[supervisord]
nodaemon=true
user=root
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid

[program:nginx]
command=nginx -g "daemon off;"
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/nginx_error.log
stdout_logfile=/var/log/supervisor/nginx_access.log

[program:registry]
command=/bin/registry serve /etc/docker/registry/config.yml
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/registry_error.log
stdout_logfile=/var/log/supervisor/registry_access.log

[program:auth-service]
command=/auth/auth-service
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/auth_error.log
stdout_logfile=/var/log/supervisor/auth_access.log
environment=DATABASE_PATH=/database/registry.db,JWT_SECRET=%(ENV_JWT_SECRET)s
EOF

# Expose ports
EXPOSE 80 443 5000 8080

# Set entrypoint
ENTRYPOINT ["/scripts/entrypoint.sh"]
CMD ["supervisord", "-c", "/etc/supervisord.conf"]