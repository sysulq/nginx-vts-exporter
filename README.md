nginx-vts-exporter
===
Simple server that scrapes Nginx vts stats and exports them via HTTP for Prometheus consumption

Dependency
---
* [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
* [Prometheus](https://prometheus.io/)
* [Golang](https://golang.org/)

Compile
---
```
go get -v ./...
go build
```

Run
---
```
nohup ./nginx-vts-exporter -nginx.scrape_uri=http://localhost/status/format/json
```

Dockerize
--

Build
```
docker build -t vts-export .
```
Run
```
docker run -ti vts-export
```

Run with args
```
docker run -ti vts-export -nginx.scrape_uri=http://localhost/status/format/json
```

