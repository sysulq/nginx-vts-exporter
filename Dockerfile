FROM        alpine:latest
MAINTAINER  Sophos <hnlq.sysu@gmail.com>

VOLUME /etc/nginx-vts-exporter/

WORKDIR /bin
COPY bin/nginx-vts-exporter /bin/
COPY bin/config.json /etc/nginx-vts-exporter/
COPY docker-entrypoint.sh /bin/
RUN chmod +x /bin/nginx-vts-exporter

ENV NGINX_HOST "http://localhost"
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

ENTRYPOINT ["docker-entrypoint.sh"]
