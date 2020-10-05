default: margarita
	docker build \
	--label org.opencontainers.image.source=https://github.com/countableset/poolside \
	-t ghcr.io/countableset/poolside:latest .

push:
	docker push ghcr.io/countableset/poolside:latest

margarita:
	cd ../margarita && make

network:
	docker network create --driver=bridge poolside-network

run:
	docker run --rm \
	--name=poolside \
	--network=poolside-network \
	-v $${HOME}/sandbox/poolside/envoy/certs/domain.crt:/etc/envoy/certs/cert.pem \
	-v $${HOME}/sandbox/poolside/envoy/certs/domain.key:/etc/envoy/certs/key.pem \
	-v $${HOME}/sandbox/poolside/envoy/certs/myCA.pem:/etc/envoy/certs/ca.pem \
	-p 443:443 \
	-p 9901:9901 \
	-p 10010:10010 \
	ghcr.io/countableset/poolside:latest
