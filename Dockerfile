FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Sophos <hnlq.sysu@gmail.com>

COPY nginx-vts-exporter  /bin/nginx-vts-exporter
COPY docker-entrypoint.sh /bin/docker-entrypoint.sh

ENV NGINX_HOST "http://localhost"
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

ENTRYPOINT [ "docker-entrypoint.sh" ]