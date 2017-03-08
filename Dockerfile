FROM alpine:3.4

# Creating directory for our application
RUN mkdir /app

# Copying entrypoint script
COPY ./docker-entrypoint.sh /app/

ENV NGIX_HOST http://localhost
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"
ENV DEFAULT_VERSION "v0.3"

EXPOSE 9913

ENTRYPOINT ["/app/docker-entrypoint.sh"]
