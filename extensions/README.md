# Extensions

The GeoIP service is extensible via a Javascript engine. The Javascript API provides a few classes that allow you to manipulate and manage IPs and IP sets.

## Anatomy

Every extension needs to be vanilla Javascript. You can choose to code your extension in Typescript, but ultimatelly it needs to compile down to one file. An example of an extension compiling down to JS can be found [here](https://github.com/wisepythagoras/geoip-service-extensions/tree/main/socks-proxy-30d) (the building is managed by [Parcel](https://parceljs.org/)). In that repository you'll also find all of the [types](https://github.com/wisepythagoras/geoip-service-extensions/blob/main/index.d.ts) for Typescript.

All extensions must have an `install` function that returns the configuration. You could also choose to bootstrap your extension in there.

``` js
function install() {
    // Do stuff here.

    return {
        version: 1,
        name: 'your_extensions_name',
        hasLookup: true,
        endpoints: [...],
        jobs: [...],
    };
}
```

The configuration is very straightforward.

* `name`: This is mandatory and should not have any spaces.
* `version`: This is not used currently, but will be in the future.
* `hasLookup`: This field is optional and if set to `true` signifies if the extension intercepts an IP lookup.
* `endpoints`: This is an array which contains all defined endpoints (see below).
* `jobs`: This is an array which contains all defined jobs (see further down).

### Endpoints

An extension could expose any amount of endpoints. The configuration for each endpoint looks as follows:

``` js
{
    method: 'GET', // Or POST, PUT, DELETE.
    handler: 'functionName',
    endpoint: '/path/to/endpoint/:param',
}
```

The `endpoint` should be defined exactly as you would define a route in [Gin](https://pkg.go.dev/github.com/gin-gonic/gin#section-readme). The `handler` member should contain the exact name of the function you intend to call when the endpoint is hit. These functions look something like this:

``` js
const functionName = (req, res) => {
    const myParam = req.param('param');

    if (isNotWhatIExpect(myParam)) {
        res.abort(500);
        return;
    }

    res.json(200, queryResult(myParam));
}
```

The [type definitions](https://github.com/wisepythagoras/geoip-service-extensions/blob/main/index.d.ts#L81-L116) will give you an idea of what is available on both `req` and `res`.

### Jobs

The configuration also allows you to add any number of cron jobs. Let's take a look at the configuration for each job.

``` js
{
    cron: '0 */2 * * *',
    job: 'jobFunctionName',
}
```

In this case we've defined a cron job which runs the function `jobFunctionName` every two hours. The function is a simple one:

``` js
const jobFunctionName = () => {
    // Do stuff here.
};
```

Your cron jobs can be used to pull data from IP lists on the web, or whatever you need.

## Javascript APIs

Coming soon.

## Full Example

``` js
const IP_LIST_DATA = 'https://some.org/some_ip_list.ipset';
const DATA_FILE = 'data.json';
let ipList = [];

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
