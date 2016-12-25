#nginx-vts-exporter

Simple server that scrapes Nginx vts stats and exports them via HTTP for Prometheus consumption

#Dependency

* [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
* [Prometheus](https://prometheus.io/)
* [Golang](https://golang.org/)

#Download
Binary can be downloaded from `bin` directory

SHA512 Sum
```
5ef43e64c47a60d44539682f645adf01beede42f855b746bb1e2b07e44e67ad279ea525d4786bc9c17a434933dfce7e48455ba05176e422c5db2831e4f284962  nginx-vts-exporter
```

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
