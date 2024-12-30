#!/bin/bash

CERT_DIR="./certs"
SLAVE_HOST=${1:-localhost}

openssl genrsa -out $CERT_DIR/slave.key 2048
openssl req -new -key $CERT_DIR/slave.key -subj "/CN=${SLAVE_HOST}" \
    -out $CERT_DIR/slave.csr
openssl x509 -req -in $CERT_DIR/slave.csr -CA $CERT_DIR/ca.crt -CAkey $CERT_DIR/ca.key \
    -CAcreateserial -out $CERT_DIR/slave.crt -days 365 -sha256

echo "Certificates have been generated for host: $SLAVE_HOST"