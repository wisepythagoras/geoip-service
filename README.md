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
  -ext-dir string
        Specify the location of the folder containing the extensions
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

The `-pub-dir` flag can be used to specify a front end application that calls all the APIs. There's an example of this in the [geoip-service-fe](https://github.com/wisepythagoras/geoip-service-fe) repository.

### Extensions

The app has an integrated extension engine which is mostly meant to be used when running it as an API server. An extension can register API endpoints, run cron jobs, and manage data on their own, which the main app can query. Below you'll find an example of an extension that queries data from a 3rd party IP list.

``` js
const IP_LIST_DATA = 'https://some.org/some_ip_list.ipset';
const DATA_FILE = 'data.json';
const ipList = [];

const listHasIP = (ip) => {
    for (let i = 0; i < ipList.length; i++) {
        if (ipList[i].contains(ip)) {
            return true;
        }
    }

    return false;
};

const lookupIP = (ip) => {
    if (listHasIP(ip)) {
        // This data will appear as `additional_info` in the API response.
        return {
            info: 'This IP was marked as XXXX',
            list: IP_LIST_DATA,
            tag: 'XXXX',
        };
    }
};

const refreshData = () => {
    // Run `fetch` to get the data and then `storage.save` to save the updated version.
};

// If you define `hasLookup` in the config then you need a function with this name.
const loadData = async () => {
    await storage.init();

    try {
        const json = await storage.read(DATA_FILE);
        const data = JSON.parse(json);

        data.forEach((ip) => {
            ipList.push(new IP(ip));
        });
    } catch (e) {
        console.log('Error:', e);
    }
};

const install = () => {
    loadData();

    return {
        version: 1,
        hasLookup: true,
        jobs: [{
            // A cron that runs every hour.
            cron: '0 * * * *',
            job: 'refreshData',
        }],
    };
};
```
