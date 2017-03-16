FROM        alpine:latest
MAINTAINER  Sophos <hnlq.sysu@gmail.com>

WORKDIR /bin
COPY bin/nginx-vts-exporter /bin/
RUN chmod +x /bin/nginx-vts-exporter

ENV NGINX_HOST "http://localhost"
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

ENTRYPOINT ["nginx-vts-exporter"]
CMD nginx-vts-exporter -nginx.scrape_uri=$NGINX_STATUS/status/format/json -telemetry.address $METRICS_ADDR -telemetry.endpoint $METRICS_ENDPOINT -metrics.namespace $METRICS_NS