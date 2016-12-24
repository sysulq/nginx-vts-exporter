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

EXPOSE 9113

ENTRYPOINT ["/app/docker-entrypoint.sh"]