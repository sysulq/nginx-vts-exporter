#nginx-vts-exporter

Simple server that scrapes Nginx vts stats and exports them via HTTP for Prometheus consumption

#Exposed metrics
##Server main
### Metrics details
Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Connections**   | `{NAMESPACE}_server_connections`| status [active, reading, writing, waiting, accepted, handled]

###Metrics output example
```
# Server Connections
nginx_server_connections{status="accepted"} 70606
```

##Server zones
### Metrics details
Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}server_requests`    | code [2xx, 3xx, 4xx, 5xx, total], host _(or domain name)_
 **Bytes**         | `{NAMESPACE}server_bytes`       | direction [in, out], host _(or domain name)_
 **Cache**         | `{NAMESPACE}server_cache`       | status [bypass, expired, hit, miss, revalidated, scarce, stale, updating], host _(or domain name)_

###Metrics output example
```
# Server Requests
nginx_server_requests{code="1xx",host="test.domain.com"} 0

# Server Bytes
nginx_server_bytes{direction="in",host="test.domain.com"} 21

# Server Cache
nginx_server_cache{host="test.domain.com",status="bypass"} 2
```

##Upstreams
### Metrics details
Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}_upstream_requests` | code [2xx, 3xx, 4xx, 5xx and total], upstream _(or upstream name)_
 **Bytes**         | `{NAMESPACE}_upstream_bytes`    | direction [in, out], upstream _(or upstream name)_
 **Response time** | `{NAMESPACE}_upstream_response` | backend (or server), in_bytes, out_bytes, upstream _(or upstream name)_
 ~~Requests/sec~~  | `NOT EXPORTED YET`              |

###Metrics output example
```
# Upstream Requests
nginx_upstream_requests{code="1xx",upstream="XXX-XXXXX-3000"} 0

# Upstream Bytes
nginx_upstream_bytes{direction="in",upstream="XXX-XXXXX-3000"} 0

# Upstream Response time
nginx_upstream_response{host="10.2.15.10:3000",in_bytes="285025.000000",out_bytes="447594.000000",upstream="XXX-XXXXX-3000"} 99
```

#Dependency

* [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
* [Prometheus](https://prometheus.io/)
* [Golang](https://golang.org/)

#Compile

```
$ ./build-binary.sh
```
This shell script above will build a temp Docker image with the binary and then
export the binary inside ./bin/ directory

#Run

```
$ nohup /bin/nginx-vts-exporter -nginx.scrape_uri=http://localhost/status/format/json
```

#Dockerized
To Dockerize this application yo need to pass two steps the build then the containerization.

## Environment variables
This image is configurable using different env variables
Variable name | Default     | Description
------------- | ----------- | --------------
NGINX_STATUS |  http://localhost/status/format/json | Nginx JSON format status page
METRICS_ENDPOINT | /metrics  | Metrics endpoint exportation URI
METRICS_ADDR | :9913 | Metrics exportation address:port
METRICS_NS | nginx | Prometheus metrics Namespaces


##Build 
```
$ ./build-binary.sh
$ docker build -t vts-export .
```

##Run
```
docker run  -ti --rm --env NGIX_HOST="http://localhost/status/format/json" --env METRICS_NS="nginx_prod1" vts-export

```
