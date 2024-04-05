#!/bin/bash

# Check for required arguments
if [ $# -ne 2 ]; then
  echo "Usage: $0 <pkcs12-file> <password>"
  exit 1
fi

pkcs12_file="$1"
password="$2"

# Extract certificate and key
openssl pkcs12 -in "$pkcs12_file" -clcerts -nokeys -out "${pkcs12_file%.p12}-cert.pem" -passin "pass:$password"
openssl pkcs12 -in "$pkcs12_file" -nocerts -out "${pkcs12_file%.p12}-key.pem" -nodes -passin "pass:$password"
