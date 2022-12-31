# GeoIP Service

A GeoIP service that can be a REST API or command line tool.

## Building

The only dependencies that this app has are [Gin](https://github.com/gin-gonic/gin), for setting up a webserver, and [govalidator](github.com/asaskevich/govalidator), for validating input.

``` sh
go build .
```

## Using

```
Usage of ./geoip-service:
  -dns-servers string
        The list of DNS servers. If not specified defaults to Cloudflare, Google, and OpenDNS
  -domain string
        A domain name
  -ip string
        An IP address
  -pub-dir string
        Specify the location of the public folder (to serve a front end)
  -serve
        Run the HTTP server
  -sip string
        The IP to serve on (127.0.0.1 will make it accessible only from localhost) (default "127.0.0.1")
  -whitelist string
        If specified, it will only allow (only used with -serve)
```

``` sh
# To query right from the command line.
./geoip-service -domain one.one.one.one
./geoip-service -ip 1.1.1.1

# To run the HTTP API.
./geoip-service -serve

# To serve on a specific iface.
./geoip-service -serve -sip 0.0.0.0

# You can also add a whitelist of IPs to allow to access the API and a custom list of
# DNS servers to query.
./geoip-service -serve -whitelist ./whitelist -sip 0.0.0.0 -dns-servers ./dns_servers
```
