#nginx-vts-exporter

Simple server that scrapes Nginx vts stats and exports them via HTTP for Prometheus consumption

#Dependency

* [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
* [Prometheus](https://prometheus.io/)
* [Golang](https://golang.org/)

#Download
Binary can be downloaded from `bin` directory.
Latest version v0.0.3

```
# SHA512 Sum
16eec84a6496529ef76a83af54f659111abecca6bcb4b2edd0b327223f93e735ae4aca2078bf4c41fded831c3d116170b277d194af64074f45992191e3a7bfb6  bin/nginx-vts-exporter
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
