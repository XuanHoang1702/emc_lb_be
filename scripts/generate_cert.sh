#!/bin/bash
set -e

SSL_DIR="$(dirname "$0")/../docker/nginx/ssl"
mkdir -p "$SSL_DIR"

if [ ! -f "$SSL_DIR/nginx-selfsigned.crt" ]; then
    echo "Generating self-signed certificate for Nginx..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$SSL_DIR/nginx-selfsigned.key" \
        -out "$SSL_DIR/nginx-selfsigned.crt" \
        -subj "/C=VN/ST=HCM/L=HCM/O=EMC/OU=IT/CN=localhost"
    echo "Certificate generated successfully."
else
    echo "Certificate already exists."
fi
