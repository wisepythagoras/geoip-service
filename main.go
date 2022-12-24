package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/maxminddb-golang"
)

var cityMmdb *maxminddb.Reader
var asnMmdb *maxminddb.Reader
var err error

// https://github.com/allegro/bigcache

func GetDomainInformation(hostname string) ([]*IPRecord, error) {
	var records []*IPRecord = []*IPRecord{}

	// Is this a valid domain name?
	if !govalidator.IsDNSName(hostname) {
		// Make sure the request is valid.
		return records, errors.New("Invalid input")
	}

	// Perform a DNS lookup.
	ips, _ := DNSLookup(hostname)

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

	if *shouldServe {
		// Run a server exposing two endpoints that are query-able.
		r := gin.Default()
		r.GET("/api/ip_address/info/:hostname", IPAddressHandler)
		r.GET("/api/domain/info/:hostname", DomainHandler)
		r.NoRoute(NotFoundHandler)

		http.ListenAndServe("127.0.0.1:8228", r)
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
