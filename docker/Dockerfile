FROM envoyproxy/envoy:v1.16-latest as envoy
####
FROM poolside/margarita as margarita
####
FROM ubuntu:20.04

ENV CONFIG_PATH=/xds/config

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y ca-certificates \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /tmp/* /var/tmp/* \
    && rm -rf /var/lib/apt/lists/*

RUN adduser --group --system envoy
RUN mkdir -p /etc/envoy/certs
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy
COPY --from=envoy /usr/local/bin/su-exec /usr/local/bin/su-exec
COPY --from=envoy /docker-entrypoint.sh /docker-entrypoint.sh

RUN mkdir -p /xds/{config,public}
COPY --from=margarita /app/app /xds/app
COPY --from=margarita /app/public/* /xds/public/

COPY start.sh /
COPY envoy.yaml /etc/envoy/envoy.yaml

CMD ["./start.sh"]