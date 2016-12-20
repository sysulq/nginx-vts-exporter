#!/bin/sh 
set -eo pipefail

# If there are any arguments then we want to run those instead
if [[ "$1" == "/nginx-vts-exporter" || -z $1 ]]; then
  exec "$@"
#$GOPATH/src/app/nginx-vts-exporter
else
  exec "/nginx-vts-exporter" "$@"
fi
