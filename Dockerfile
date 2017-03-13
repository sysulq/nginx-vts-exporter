FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Sophos <hnlq.sysu@gmail.com>

COPY nginx_vts_exporter /bin/nginx_vts_exporter

EXPOSE      9913
ENTRYPOINT  [ "/bin/nginx_vts_exporter" ]