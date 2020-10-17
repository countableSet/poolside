#!/usr/bin/env sh
set -e

cert_filename="/etc/envoy/certs/cert.pem"
key_filename="/etc/envoy/certs/key.pem"
ca_filename="/etc/envoy/certs/ca.pem"
ca_key_filename="/etc/envoy/certs/ca.key"
domain_filename="/etc/envoy/certs/domain.ext"
# Generate certificates if none are mounted needed
if [ ! -f "$cert_filename" ] || [ ! -f "$key_filename" ]; then
  set -x
  if [ ! -f "$domain_filename" ]; then
    echo "authorityKeyIdentifier=keyid,issuer\nbasicConstraints=CA:FALSE\nkeyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment\nsubjectAltName = @alt_names\n\n[alt_names]\nDNS.1 = *.poolside.dev" > $domain_filename
  fi
  if [ ! -f "$ca_filename" ] || [ ! -f "$ca_key_filename"]; then
    # generate root ca private key
    openssl genrsa -out $ca_key_filename 2048
    # generate root certificate
    openssl req -x509 -new -nodes -key $ca_key_filename -sha256 -days 356 -out $ca_filename \
      -subj "/C=US/ST=California/L=AnyCity/O=Poolside/OU=Org/CN=poolside.dev"
  fi
  # generate private key for domain
  openssl genrsa -out $key_filename 4096
  # create csr?
  openssl req -new -key $key_filename -out domain.csr \
    -subj "/C=US/ST=California/L=AnyCity/O=Poolside/OU=Org/CN=poolside.dev"
  # cert certificate
  openssl x509 -req -in domain.csr -CA $ca_filename -CAkey $ca_key_filename -CAcreateserial \
    -out $cert_filename -days 365 -sha256 -extfile $domain_filename
  set +x
fi

# Start the first process (envoy)
./docker-entrypoint.sh /usr/local/bin/envoy -c /etc/envoy/envoy.yaml &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start enovy process: $status"
  exit $status
fi

# Start the second process (xds)
cd /xds
./app &
cd /
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start xds process: $status"
  exit $status
fi

# Naive check runs checks once every 30 seconds to see if either of the 
# processes exited. This illustrates part of the heavy lifting you need 
# to do if you want to run more than one service in a container. The 
# container exits with an error if it detects that either of the processes 
# has exited. Otherwise it loops forever, waking up every 30 seconds.

while sleep 30; do
  pgrep -a envoy
  PROCESS_1_STATUS=$?
  pgrep -a app
  PROCESS_2_STATUS=$?
  # If the greps above find anything, they exit with 0 status
  # If they are not both 0, then something is wrong
  if [ $PROCESS_1_STATUS -ne 0 -o $PROCESS_2_STATUS -ne 0 ]; then
    echo "One of the processes has already exited."
    exit 1
  fi
done
