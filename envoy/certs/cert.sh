#!/bin/bash
set -x

# generate private key for domain
openssl genrsa -out domain.key 4096

# create csr?
openssl req -new -key domain.key -out domain.csr \
    -subj "/C=US/ST=California/L=AnyCity/O=Poolside/OU=Org/CN=*.local.bimmer-tech.com"

# cert certificate
openssl x509 -req -in domain.csr -CA myCA.pem -CAkey myCA.key -CAcreateserial \
    -out domain.crt -days 365 -sha256 -extfile domain.ext
