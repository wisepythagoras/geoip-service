# GeoIP Service

A GeoIP service that can be a REST API or command line tool.

## Building

``` sh
go get github.com/oschwald/maxminddb-golang \
    github.com/gorilla/mux \
    github.com/asaskevich/govalidator

go build .
```

## Using

```
Usage of ./geoip-service:
  -domain string
    	A domain name
  -ip string
    	An IP address
  -serve
    	Run the HTTP server
```

``` sh
# To query right from the command line.
./geoip-service -domain one.one.one.one
./geoip-service -ip 1.1.1.1

# To run the HTTP API.
./geoip-service -serve
```
