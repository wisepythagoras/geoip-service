package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/maxminddb-golang"
)

var cityMmdb *maxminddb.Reader
var asnMmdb *maxminddb.Reader
var err error
var whiteListedIPRanges []*net.IPNet
var whiteListedIPs []net.IP
var hasWhitelist = false
var dnsServerList = []string{}

func middleware(c *gin.Context) {
	// If there was no whitelist specified, then we can proceed.
	if !hasWhitelist {
		c.Next()
		return
	}

	clientIP := net.ParseIP(c.ClientIP())

	if val := c.GetHeader("True-Client-IP"); len(val) > 0 {
		clientIP = net.ParseIP(val)
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

func GetDomainInformation(hostname string) ([]*IPRecord, error) {
	var records []*IPRecord = []*IPRecord{}

	// Is this a valid domain name?
	if !govalidator.IsDNSName(hostname) {
		// Make sure the request is valid.
		return records, errors.New("invalid input")
	}

	// Perform a DNS lookup.
	ips, _ := DNSLookup(hostname, dnsServerList)

	for i := 0; i < len(ips); i++ {
		// Get the information on the current IP.
		info, err := GetIPInformation(ips[i].String())

		if err != nil {
			continue
		}

		// Append the record to the array.
		records = append(records, info)
	}

	return records, nil
}

func GetIPInformation(hostname string) (*IPRecord, error) {
	// If you are using strings that may be invalid, check that ip is not nil.
	ip := net.ParseIP(hostname)

	// Create an instance of the IP record.
	rec := &IPRecord{}

	// Lookup the IP details from the city database.
	err = cityMmdb.Lookup(ip, &rec)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Lookup the IP details from the ASN database.
	err = asnMmdb.Lookup(ip, &rec)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	rec.IPAddress = hostname

	return rec, nil
}

func IPAddressHandler(c *gin.Context) {
	hostname := c.Param("hostname")
	response := &ApiResponse{}
	response.Record = nil

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
	response.Record, err = GetIPInformation(hostname)

	if err != nil {
		response.Status = err.Error()
	}

	c.JSON(200, response)
}

func DomainHandler(c *gin.Context) {
	hostname := c.Param("hostname")
	response := &MultiApiResponse{}
	response.Records, err = GetDomainInformation(hostname)

	if err == nil {
		response.Success = true
		response.Status = "Retrieved"
	} else {
		response.Success = false
		response.Status = err.Error()
	}

	c.JSON(200, response)
}

func NotFoundHandler(c *gin.Context) {
	c.Status(404)
}

func main() {
	domainPtr := flag.String("domain", "", "A domain name")
	ipPtr := flag.String("ip", "", "An IP address")
	shouldServe := flag.Bool("serve", false, "Run the HTTP server")
	serveIP := flag.String("sip", "127.0.0.1", "The IP to serve on (127.0.0.1 will make it accessible only from localhost)")
	whitelist := flag.String("whitelist", "", "If specified, it will only allow (only used with -serve)")
	dnsServers := flag.String("dns-servers", "", "The list of DNS servers. If not specified defaults to Cloudflare, Google, and OpenDNS")

	flag.Parse()

	// Open the city database.
	cityMmdb, err = OpenCityDB()

	if err != nil {
		log.Fatal(err)
	}

	// Open the ASN database.
	asnMmdb, err = OpenASNDB()

	if err != nil {
		log.Fatal(err)
	}

	defer cityMmdb.Close()
	defer asnMmdb.Close()

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
		r.GET("/api/ip_address/info/:hostname", IPAddressHandler)
		r.GET("/api/domain/info/:hostname", DomainHandler)
		r.NoRoute(NotFoundHandler)

		http.ListenAndServe(fmt.Sprintf("%s:8228", *serveIP), r)
	} else if *domainPtr != "" {
		// Grab the domain information.
		recs, _ := GetDomainInformation(*domainPtr)
		obj, _ := json.Marshal(recs)
		fmt.Println(string(obj))
	} else if *ipPtr != "" {
		// Grab the information about the sole IP address.
		rec, _ := GetIPInformation(*ipPtr)
		obj, _ := json.Marshal(rec)
		fmt.Println(string(obj))
	} else {
		fmt.Println("Nothing queried")
	}
}
