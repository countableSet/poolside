#!/bin/bash
set -x

# generate root ca private key
openssl genrsa -out myCA.key 2048

# generate root certificate
openssl req -x509 -new -nodes -key myCA.key -sha256 -days 356 -out myCA.pem \
    -subj "/C=US/ST=California/L=AnyCity/O=Poolside/OU=Org/CN=*.poolside.dev"
