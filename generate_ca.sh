#!/bin/bash

CERT_DIR="./certs"
mkdir -p $CERT_DIR

openssl genrsa -out $CERT_DIR/ca.key 2048
openssl req -x509 -new -nodes -key $CERT_DIR/ca.key -sha256 -days 365 \
    -subj "/CN=BloaderCA" -out $CERT_DIR/ca.crt

echo "Generated CA certificate"