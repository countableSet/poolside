# poolside 

Configurable auto-reload proxy to put services behind domain names for better cookie management.

#### getting started

```
docker pull ghcr.io/countableset/poolside:latest
docker network create --driver=bridge poolside-network
docker run \
	--name=poolside \
	--network=poolside-network \
	-v $HOME/.poolside/certs/:/etc/envoy/certs/ \
	-v $HOME/.poolside/config.json:/xds/config.json \
	-p 443:443 \
	-p 9901:9901 \
	-p 10010:10010 \
	ghcr.io/countableset/poolside:latest
```

Open [http://localhost:10010](http://localhost:10010) for configuration ui.

Connect to docker-compose containers to the poolside network to allow for forwarding:
```
networks:
  default:
    external:
      name: poolside-network
```

Connect standalone container to the poolside network to allow for forwarding:
```
--network=poolside-network
```