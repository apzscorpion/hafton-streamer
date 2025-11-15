#!/bin/bash

# SSL Setup Script for Let's Encrypt
# Run this script after setting up your domain DNS

set -e

DOMAIN="${1:-yourdomain.com}"
EMAIL="${2:-admin@${DOMAIN}}"

echo "Setting up SSL certificate for ${DOMAIN}"

# Install certbot if not already installed
if ! command -v certbot &> /dev/null; then
    echo "Installing certbot..."
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
fi

# Stop nginx temporarily for standalone mode (if not using nginx plugin)
# systemctl stop nginx

# Obtain certificate
echo "Obtaining SSL certificate..."
certbot certonly --standalone -d "${DOMAIN}" -d "www.${DOMAIN}" --email "${EMAIL}" --agree-tos --non-interactive

# Or use nginx plugin (if nginx is already configured)
# certbot --nginx -d "${DOMAIN}" -d "www.${DOMAIN}" --email "${EMAIL}" --agree-tos --non-interactive

echo "SSL certificate obtained successfully!"
echo "Update nginx/streaming.conf with your domain name"
echo "Then restart nginx: systemctl restart nginx"

# Set up auto-renewal
echo "Setting up auto-renewal..."
systemctl enable certbot.timer
systemctl start certbot.timer

echo "SSL setup complete!"

