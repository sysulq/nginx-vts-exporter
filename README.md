#nginx-vts-exporter
===
Simple server that scrapes Nginx vts stats and exports them via HTTP for Prometheus consumption

#Dependency
---
* [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
* [Prometheus](https://prometheus.io/)
* [Golang](https://golang.org/)

#Compile
---
```
$ ./build-binary.sh
```
This shell script above will build a temp Docker image with the binary and then
export the binary inside ./bin/ directory

#Run
---
```
$ nohup /bin/nginx-vts-exporter -nginx.scrape_uri=http://localhost/status/format/json
```

#Dockerize
--

##Build
```
$ ./build-binary.sh
$ docker build -t vts-export .
```
##Run
```
docker run -ti vts-export
```

##Run with args
```
docker run -ti vts-export -nginx.scrape_uri=http://localhost/status/format/json
```
