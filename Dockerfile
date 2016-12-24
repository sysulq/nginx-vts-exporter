FROM alpine:3.4

# Creating directory for our application
RUN mkdir /app

# Copying entrypoint script
COPY ./docker-entrypoint.sh /app/
# Copying VTS Metrics exporter binary
COPY bin/nginx-vts-exporter /app/

ENV NGIX_HOST http://localhost
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

EXPOSE 9913

ENTRYPOINT ["/app/docker-entrypoint.sh"]
