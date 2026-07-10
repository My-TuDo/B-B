#!/bin/sh

SSL_DIR=/etc/nginx/ssl

if [ ! -f "$SSL_DIR/server.crt" ]; then
    mkdir -p "$SSL_DIR"
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$SSL_DIR/server.key" \
        -out "$SSL_DIR/server.crt" \
        -subj "/C=CN/ST=Shanghai/L=Shanghai/O=B-B/OU=Dev/CN=localhost"
    echo "Self-signed certificate generated"
fi
