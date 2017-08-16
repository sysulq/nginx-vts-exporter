# nginx-vts-exporter

[![Build Status](https://travis-ci.org/hnlq715/nginx-vts-exporter.svg?branch=master)](https://travis-ci.org/hnlq715/nginx-vts-exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/sophos/nginx-vts-exporter.svg)](https://hub.docker.com/r/sophos/nginx-vts-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/hnlq715/nginx-vts-exporter)](https://goreportcard.com/report/github.com/hnlq715/nginx-vts-exporter)
[![Gratipay User](https://img.shields.io/gratipay/user/hnlq715.svg)](https://gratipay.com/hnlq715/)

Simple server that scrapes Nginx [vts](https://github.com/vozlt/nginx-module-vts) stats and exports them via HTTP for Prometheus consumption

## Table of Contents
* [Dependency](#dependency)
* [Download](#download)
* [Compile](#compile)
  * [build binary](#build-binary)
  * [build docker image](#build-docker-image)
* [Run](#run)
  * [run binary](#run-binary)
  * [run docker image](#run-docker-image)
* [Environment variables](#environment-variables)
* [Metrics](#metrics)
  * [Server main](#server-main)
  * [Server zones](#server-zones)
  * [Upstreams](#upstreams)

## Dependency

* [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
* [Prometheus](https://prometheus.io/)
* [Golang](https://golang.org/)

## Download

Binary can be downloaded from [Releases](https://github.com/hnlq715/nginx-vts-exporter/releases) page.

## Compile

### build binary

``` shell
make
```

### build docker image
``` shell
make docker
```

## Docker Hub Image
``` shell
docker pull sophos/nginx-vts-exporter:latest
```
It can be used directly instead of having to build the image yourself.
([Docker Hub sophos/nginx-vts-exporter](https://hub.docker.com/r/sophos/nginx-vts-exporter/))

## Run

### run binary
``` shell
nohup /bin/nginx-vts-exporter -nginx.scrape_uri=http://localhost/status/format/json
```

### run docker
```
docker run  -ti --rm --env NGINX_HOST="http://localhost/status/format/json" sophos/nginx-vts-exporter
```

## Environment variables

This image is configurable using different env variables

Variable name | Default     | Description
------------- | ----------- | --------------
NGINX_STATUS |  http://localhost/status/format/json | Nginx JSON format status page
METRICS_ENDPOINT | /metrics  | Metrics endpoint exportation URI
METRICS_ADDR | :9913 | Metrics exportation address:port
METRICS_NS | nginx | Prometheus metrics Namespaces

## Metrics

Documents about exposed Prometheus metrics.

For details on the underlying metrics please see [nginx-module-vts](https://github.com/vozlt/nginx-module-vts#json-used-by-status)

For grafana dashboard please see [nginx-vts-exporter dashboard](https://grafana.com/dashboards/2949)

### Server main

**Metrics details**

Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Connections**   | `{NAMESPACE}_server_connections`| status [active, reading, writing, waiting, accepted, handled]

**Metrics output example**

``` txt
# Server Connections
nginx_server_connections{status="accepted"} 70606
```

### Server zones

**Metrics details**

Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}_server_requests`    | code [2xx, 3xx, 4xx, 5xx, total], host _(or domain name)_
 **Bytes**         | `{NAMESPACE}_server_bytes`       | direction [in, out], host _(or domain name)_
 **Cache**         | `{NAMESPACE}_server_cache`       | status [bypass, expired, hit, miss, revalidated, scarce, stale, updating], host _(or domain name)_

**Metrics output example**

``` txt
# Server Requests
nginx_server_requests{code="1xx",host="test.domain.com"} 0

# Server Bytes
nginx_server_bytes{direction="in",host="test.domain.com"} 21

# Server Cache
nginx_server_cache{host="test.domain.com",status="bypass"} 2
```

### Upstreams

**Metrics details**

Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}_upstream_requests` | code [2xx, 3xx, 4xx, 5xx and total], upstream _(or upstream name)_
 **Bytes**         | `{NAMESPACE}_upstream_bytes`    | direction [in, out], upstream _(or upstream name)_
 **Response time** | `{NAMESPACE}_upstream_response` | backend (or server), in_bytes, out_bytes, upstream _(or upstream name)_

**Metrics output example**

``` txt
# Upstream Requests
nginx_upstream_requests{code="1xx",upstream="XXX-XXXXX-3000"} 0

# Upstream Bytes
nginx_upstream_bytes{direction="in",upstream="XXX-XXXXX-3000"} 0

# Upstream Response time
nginx_upstream_response{backend="10.2.15.10:3000",upstream="XXX-XXXXX-3000"} 99
```
