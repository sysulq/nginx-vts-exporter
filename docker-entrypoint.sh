#!/bin/sh
set -eo pipefail
binary="/app/nginx-vts-exporter"
default_status="$NGIX_HOST/status/format/json"
NGINX_STATUS=${NGINX_STATUS:-$default_status}
METRICS_NS=${METRICS_NS:-$DEFAULT_METRICS_NS}

# If there are any arguments then we want to run those instead
#if [[ "$1" == "$binary" || -z $1 ]]; then
#  exec "$@"
#else
#  echo "Running the default"
#echo "[$0] - Nginx scrape host --> [$NGINX_STATUS]"
#echo "[$0] - Metrics Address   --> [$METRICS_ADDR]"
#echo "[$0] - Metrics Endpoint  --> [$METRICS_ENDPOINT]"
#echo "[$0] - Metrics Namespace  --> [$METRICS_NS]"
#echo "[$0] - Running metrics nginx-vts-exporter"
exec "$binary" -nginx.scrape_uri=$NGINX_STATUS -telemetry.address $METRICS_ADDR -telemetry.endpoint $METRICS_ENDPOINT -metrics.namespace $METRICS_NS
#fi
