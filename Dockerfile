FROM        quay.io/prometheus/busybox:latest
LABEL       Sophos <hnlq.sysu@gmail.com>

COPY ./dist/nginx-vtx-exporter_linux_amd64_v1/nginx-vtx-exporter  /bin/nginx-vts-exporter
COPY docker-entrypoint.sh /bin/docker-entrypoint.sh

ENV NGINX_HOST "http://localhost"
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

ENTRYPOINT [ "docker-entrypoint.sh" ]