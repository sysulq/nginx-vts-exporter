FROM golang:alpine

#RUN apt-get update &&\
#    rm -rf /var/lib/apt/lists
RUN apk --no-cache --update add git ca-certificates
WORKDIR $GOPATH/src/app/
ADD . . 
RUN go get -v
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nginx-vts-exporter .
RUN mv $GOPATH/src/app/nginx-vts-exporter /
COPY ./docker-entrypoint.sh /

EXPOSE 9113

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/nginx-vts-exporter"]
#, "-nginx.scrape_uri=http://localhost/status/format/json"]

#CMD ["$GOPATH/src/app/nginx-vts-exporter"] 
