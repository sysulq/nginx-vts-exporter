
FROM golang:alpine

RUN apk --no-cache --update add git ca-certificates
WORKDIR $GOPATH/src/app/
ADD . .
RUN go get -v
RUN mkdir /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nginx-vts-exporter .
RUN mv $GOPATH/src/app/nginx-vts-exporter /app/

COPY ./docker-entrypoint.sh /app/
ENV NGIX_HOST http://localhost
ENV METRICS_ENDPOINT "/metrics"
ENV METRICS_ADDR ":9913"
ENV DEFAULT_METRICS_NS "nginx"

ENTRYPOINT ["docker-entrypoint.sh"]
