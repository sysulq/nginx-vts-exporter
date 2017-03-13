FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Sophos <hnlq.sysu@gmail.com>

COPY bin/nginx-vts-exporter.tar.gz /bin/
RUN tar xf /bin/nginx_vts_exporter.tar.gz /bin/nginx_vts_exporter
COPY docker-entrypoint.sh /bin/docker-entrypoint.sh

ENV NGIX_HOST "http://localhost"
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

ENTRYPOINT ["/bin/docker-entrypoint.sh"]