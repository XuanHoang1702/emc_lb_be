#!/bin/bash
set -e

echo "Setting up UFW Firewall Rules for EMC LB Server..."

# Ensure UFW is installed
if ! command -v ufw >/dev/null 2>&1; then
    echo "UFW is not installed. Please install it first."
    echo "Ubuntu/Debian: sudo apt-get update && sudo apt-get install ufw"
    exit 1
fi

# Set defaults
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (Port 22) - Important: change port if you use a custom SSH port
echo "Allowing SSH..."
sudo ufw allow 22/tcp

# Allow HTTP and HTTPS for Nginx
echo "Allowing HTTP/HTTPS..."
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow WireGuard VPN
echo "Allowing WireGuard..."
sudo ufw allow 51820/udp

# DOCKER RULE: Docker often bypasses UFW by manipulating iptables directly.
# To properly restrict public access to ports exposed via Docker (if any were missed),
# you typically need to add rules to the DOCKER-USER chain.
# However, in this architecture, we have removed public port bindings in docker-compose.yml 
# (e.g. 9090, 3000, 9273), so Docker will not expose them to 0.0.0.0 anyway.

echo "UFW rules applied. Here is the planned status:"
sudo ufw show added

echo ""
echo "To enable the firewall, run: sudo ufw enable"
echo "WARNING: Ensure your SSH port is allowed before enabling, otherwise you will lose access!"
