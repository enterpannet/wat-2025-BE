#!/bin/bash

# Script to update registration-api.service on server
# Usage: ./update-service.sh

SERVICE_FILE="/etc/systemd/system/registration-api.service"
BACKUP_FILE="/etc/systemd/system/registration-api.service.backup.$(date +%Y%m%d_%H%M%S)"

echo "Updating registration-api.service..."

# Create backup
if [ -f "$SERVICE_FILE" ]; then
    echo "Creating backup: $BACKUP_FILE"
    sudo cp "$SERVICE_FILE" "$BACKUP_FILE"
fi

# Stop service
echo "Stopping service..."
sudo systemctl stop registration-api || true

# Update service file
echo "Updating service file..."
sudo tee "$SERVICE_FILE" > /dev/null <<'EOF'
[Unit]
Description=Registration System API (mostdata.site)
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/registration-system/backend
ExecStart=/var/www/registration-system/backend/registration-api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Environment variables
Environment="PORT=3000"
Environment="CORS_ORIGINS=https://mostdata.site,https://www.mostdata.site"
Environment="HTTPS_ENABLED=true"
Environment="ENVIRONMENT=production"
Environment="CLOUDINARY_CLOUD_NAME=dq3yssqpv"
Environment="CLOUDINARY_API_KEY=913319833464127"
Environment="CLOUDINARY_API_SECRET=qPdBLJJaBmiq5tDT0YusXkRc4Hw"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=registration_user"
Environment="DB_PASSWORD=your_secure_password"
Environment="DB_NAME=registration_db"
Environment="DB_SSL_MODE=disable"

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Resource limits
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable registration-api

# Start service
echo "Starting service..."
sudo systemctl start registration-api

# Wait a bit
sleep 2

# Check status
echo "Checking service status..."
sudo systemctl status registration-api --no-pager

echo "✅ Service updated successfully!"

