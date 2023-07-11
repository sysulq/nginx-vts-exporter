# nginx-vts-exporter

[![Go](https://github.com/hnlq715/nginx-vts-exporter/actions/workflows/go.yml/badge.svg)](https://github.com/hnlq715/nginx-vts-exporter/actions/workflows/go.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/sophos/nginx-vts-exporter.svg)](https://hub.docker.com/r/sophos/nginx-vts-exporter)
[![Github All Releases](https://img.shields.io/github/downloads/hnlq715/nginx-vts-exporter/total.svg)](https://github.com/hnlq715/nginx-vts-exporter)
[![GitHub release](https://img.shields.io/github/release/hnlq715/nginx-vts-exporter.svg)](https://github.com/hnlq715/nginx-vts-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/hnlq715/nginx-vts-exporter)](https://goreportcard.com/report/github.com/hnlq715/nginx-vts-exporter)

Simple server that scrapes Nginx [vts](https://github.com/vozlt/nginx-module-vts) stats and exports them via HTTP for Prometheus consumption

To support time related histogram metrics, please refer to [hnlq715/nginx-prometheus-metrics](https://github.com/hnlq715/nginx-prometheus-metrics) or [#43](https://github.com/hnlq715/nginx-vts-exporter/issues/43).

## ANN

It's hard to say that this project is **not maintained** any longer, and it is recommended to use [nginx-vtx-module](https://github.com/vozlt/nginx-module-vts) instead, which supports multiple vhost_traffic_status_display_format, like `<json|html|jsonp|prometheus>`.

Hope you guys enjoy it, and thanks for all the contributors and the issue finders. 😃

## Table of Contents
- [nginx-vts-exporter](#nginx-vts-exporter)
  - [ANN](#ann)
  - [Table of Contents](#table-of-contents)
  - [Dependency](#dependency)
  - [Download](#download)
  - [Compile](#compile)
    - [build binary](#build-binary)
    - [build RPM package](#build-rpm-package)
    - [build docker image](#build-docker-image)
  - [Docker Hub Image](#docker-hub-image)
  - [Run](#run)
    - [run binary](#run-binary)
    - [run docker](#run-docker)
  - [Environment variables](#environment-variables)
  - [Metrics](#metrics)
    - [Server main](#server-main)
    - [Server zones](#server-zones)
    - [Filter zones](#filter-zones)
    - [Upstreams](#upstreams)

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

### build RPM package
``` shell
make rpm
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
docker run  -ti --rm --env NGINX_STATUS="http://localhost/status/format/json" sophos/nginx-vts-exporter
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
 **Info**          | `{NAMESPACE}_server_info`       | hostName, nginxVersion, uptimeSec |
 **Connections**   | `{NAMESPACE}_server_connections`| status [active, reading, writing, waiting, accepted, handled]

**Metrics output example**

``` txt
# Server Info
nginx_server_info{hostName="localhost", nginxVersion="1.11.1"} 9527
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

### Filter zones

**Metrics details**

Nginx data         | Name                              | Exposed informations
------------------ | --------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}_filter_requests`     | code [2xx, 3xx, 4xx, 5xx and total], filter, filter name
 **Bytes**         | `{NAMESPACE}_filter_bytes`        | direction [in, out], filter, filter name
 **Response time** | `{NAMESPACE}_filter_responseMsec` | filter, filter name

**Metrics output example**

``` txt
# Filter Requests
nginx_upstream_requests{code="1xx", filter="country", filterName="BY"} 0

# Filter Bytes
nginx_upstream_bytes{direction="in", filter="country", filterName="BY"} 0

# Filter Response time
nginx_upstream_responseMsec{filter="country", filterName="BY"} 99
```


### Upstreams

**Metrics details**

Nginx data         | Name                                | Exposed informations
------------------ | ----------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}_upstream_requests`     | code [2xx, 3xx, 4xx, 5xx and total], upstream _(or upstream name)_
 **Bytes**         | `{NAMESPACE}_upstream_bytes`        | direction [in, out], upstream _(or upstream name)_
 **Response time** | `{NAMESPACE}_upstream_responseMsec` | backend (or server), in_bytes, out_bytes, upstream _(or upstream name)_

**Metrics output example**

``` txt
# Upstream Requests
nginx_upstream_requests{code="1xx",upstream="XXX-XXXXX-3000"} 0

# Upstream Bytes
nginx_upstream_bytes{direction="in",upstream="XXX-XXXXX-3000"} 0

# Upstream Response time
nginx_upstream_responseMsec{backend="10.2.15.10:3000",upstream="XXX-XXXXX-3000"} 99
```
