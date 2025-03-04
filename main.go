package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/wisepythagoras/geoip-service/crypto"
	"github.com/wisepythagoras/geoip-service/db"
	"github.com/wisepythagoras/geoip-service/dns"
	"github.com/wisepythagoras/geoip-service/extension"
	"github.com/wisepythagoras/geoip-service/types"
)

var database *db.DB
var err error
var whiteListedIPRanges []*net.IPNet
var whiteListedIPs []net.IP
var hasWhitelist = false
var dnsServerList = []string{}
var extensions []*extension.Extension
var appAPIKey string

func middleware(c *gin.Context) {
	apiKey := c.GetHeader("X-AUTH-TOKEN")
	method := c.Request.Method
	requiresAPIKey := method == "POST" || method == "PUT" || method == "DELETE"

	// If there was no whitelist specified, then we can proceed.
	if !hasWhitelist {
		if (len(apiKey) == 0 || apiKey != appAPIKey) && requiresAPIKey {
			c.AbortWithStatus(401)
			return
		}

		c.Next()
		return
	}

	clientIP := net.ParseIP(c.ClientIP())

	if val := c.GetHeader("True-Client-IP"); len(val) > 0 {
		clientIP = net.ParseIP(val)
	}

	if (len(apiKey) == 0 || apiKey != appAPIKey) && requiresAPIKey {
		c.AbortWithStatus(401)
		return
	}

	// Otherwise we need to check both list of IPs and IP ranges.
	if sliceContains(whiteListedIPs, clientIP) {
		c.Next()
		return
	}

	for _, ipRange := range whiteListedIPRanges {
		if ipRange.Contains(clientIP) {
			c.Next()
			return
		}
	}

	// If the client's IP address was not found in the whitelisted IPs, then we should deny access.
	c.AbortWithStatus(400)
}

// https://github.com/allegro/bigcache

func IPAddressHandler(c *gin.Context) {
	hostname := c.Param("hostname")
	clientIPStr := c.ClientIP()
	clientIP := net.ParseIP(clientIPStr)
	response := &types.ApiResponse{}
	response.Data = nil

	// Is this a valid IP address?
	if !IsValidIP(hostname) {
		response.Success = false
		response.Status = "Invalid input"

		c.JSON(500, response)

		return
	}

	response.Success = true
	response.Status = "Retrieved"

	// Get the IP information for this.
	response.Data, err = database.GetIPInformation(hostname, &clientIP)

	if err != nil {
		response.Status = err.Error()
	}

	c.JSON(200, response)
}

func FastDomainHandler(c *gin.Context) {
	hostname := c.Param("hostname")
	clientIPStr := c.ClientIP()
	clientIP := net.ParseIP(clientIPStr)
	response := &types.ApiResponse{}
	response.Data, err = database.GetDomainInformation(hostname, dnsServerList, &clientIP)

	if err == nil {
		response.Success = true
		response.Status = "Retrieved"
	} else {
		response.Success = false
		response.Status = err.Error()
	}

	c.JSON(200, response)
}

func DomainHandler(c *gin.Context) {
	hostname := c.Param("hostname")
	clientIPStr := c.ClientIP()
	clientIP := net.ParseIP(clientIPStr)
	response := &types.ApiResponse{}
	response.Data, err = database.GetDomainInfoFromDNS(hostname, dnsServerList, dns.DNSALookup, &clientIP)

	if err == nil {
		response.Success = true
		response.Status = "Retrieved"
	} else {
		response.Success = false
		response.Status = err.Error()
	}

	c.JSON(200, response)
}

func DNSServers(c *gin.Context) {
	response := &types.ApiResponse{
		Success: true,
	}

	if len(dnsServerList) > 0 {
		response.Data = dnsServerList
	} else {
		response.Data = dns.DefaultDNSServers
	}

	c.JSON(200, response)
}

func main() {
	domainPtr := flag.String("domain", "", "A domain name")
	ipPtr := flag.String("ip", "", "An IP address")
	shouldServe := flag.Bool("serve", false, "Run the HTTP server")
	serveIP := flag.String("sip", "127.0.0.1", "The IP to serve on (127.0.0.1 will make it accessible only from localhost)")
	whitelist := flag.String("whitelist", "", "If specified, it will only allow access to the IPs in the list (only used with -serve)")
	dnsServers := flag.String("dns-servers", "", "The list of DNS servers. If not specified defaults to Cloudflare, Google, and OpenDNS")
	publicFolder := flag.String("pub-dir", "", "Specify the location of the public folder (to serve a front end)")
	extFolder := flag.String("ext-dir", "", "Specify the location of the folder containing the extensions")
	apiKey := flag.String("api-key", "", "Specify an API key to protect your instance (it will be generated if you don't specify one)")

	flag.Parse()

	if len(*extFolder) > 0 {
		extensions, err = parseExtensions(*extFolder)

		if err != nil {
			fmt.Println("Load error:", err)
			os.Exit(1)
		}

		for _, e := range extensions {
			err = e.Init()

			if err != nil {
				fmt.Println("Init error:", err)
				os.Exit(1)
			}
		}
	}

	// Open the city database.
	database = &db.DB{Extensions: extensions}
	err := database.Open()

	if err != nil {
		log.Fatal(err)
	}

	if len(*dnsServers) > 0 {
		file, err := os.Open(*dnsServers)

		if err != nil {
			fmt.Println("Unable to open the specified whitelist file")
			os.Exit(1)
		}

		defer file.Close()
		dnsServerList, err = ParseDNSServerList(file)

		if err != nil {
			fmt.Println("Error while reading the DNS server list file")
			os.Exit(1)
		}
	}

	if *shouldServe {
		if len(*apiKey) > 0 {
			appAPIKey = *apiKey
		} else {
			randBytes, err := crypto.GenRandomBytes(32, time.Now().Unix())

			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			hashBytes, err := crypto.GetSHA256Hash(randBytes)

			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			appAPIKey = crypto.ByteArrayToHex(hashBytes)
		}

		fmt.Println("API key:", appAPIKey)
		fmt.Println("This API key should be used to access any non-GET endpoint")

		if len(*whitelist) > 0 {
			file, err := os.Open(*whitelist)

			if err != nil {
				fmt.Println("Unable to open the specified whitelist file")
				os.Exit(1)
			}

			defer file.Close()
			whiteListedIPRanges, whiteListedIPs, err = ParseIPList(file)

			if err != nil {
				fmt.Println("Error while parsing the whitelist", err)
				os.Exit(1)
			}

			hasWhitelist = true
		}

		// Run a server exposing two endpoints that are query-able.
		r := gin.Default()

		r.Use(middleware)

		r.NoRoute(func(c *gin.Context) {
			if len(*publicFolder) > 0 {
				c.File(path.Join(*publicFolder, "index.html"))
			}

			c.Status(http.StatusNotFound)
		})

		if len(*publicFolder) > 0 {
			if !fileExists(*publicFolder) {
				fmt.Println("The provided public folder doesn't exist")
				os.Exit(1)
			}

			configPath := path.Join(*publicFolder, "conf.json")
			paths := make(map[string]bool)
			paths["/"] = true

			if fileExists(configPath) {
				config, err := ReadConfig(fmt.Sprintf("%s/conf.json", *publicFolder))

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				for _, endpoint := range config.Views {
					paths[endpoint] = true
				}
			}

			// Add a public folder, if one was specified. This is available so that a you can run
			// a front end application, instead of using it just as an API.
			r.Use(static.Serve("/", static.LocalFile(*publicFolder, false)))

			for endpoint, _ := range paths {
				r.GET(endpoint, func(c *gin.Context) {
					c.File(path.Join(*publicFolder, "index.html"))
				})
			}
		}

		r.GET("/api/ip_address/info/:hostname", IPAddressHandler)
		r.GET("/api/domain/fast_info/:hostname", FastDomainHandler)
		r.GET("/api/domain/info/:hostname", DomainHandler)
		r.GET("/api/dns_servers", DNSServers)

		// Register any endpoint extensions.
		for _, ext := range extensions {
			if !ext.IsEndpointExtension() {
				continue
			}

			ext.RegisterEndpoints(r)
		}

		http.ListenAndServe(fmt.Sprintf("%s:8228", *serveIP), r)
	} else if *domainPtr != "" {
		// Grab the domain information.
		recs, _ := database.GetDomainInformation(*domainPtr, dnsServerList, nil)
		obj, _ := json.Marshal(recs)
		fmt.Println(string(obj))
	} else if *ipPtr != "" {
		// Grab the information about the sole IP address.
		rec, _ := database.GetIPInformation(*ipPtr, nil)
		obj, _ := json.Marshal(rec)
		fmt.Println(string(obj))
	} else {
		fmt.Println("Nothing queried")
	}
}
