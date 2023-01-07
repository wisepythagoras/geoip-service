package db

import (
	"errors"
	"log"
	"net"

	"github.com/asaskevich/govalidator"
	"github.com/oschwald/maxminddb-golang"
	"github.com/wisepythagoras/geoip-service/dns"
	"github.com/wisepythagoras/geoip-service/extension"
	"github.com/wisepythagoras/geoip-service/types"
)

type DB struct {
	cityMmdb   *maxminddb.Reader
	asnMmdb    *maxminddb.Reader
	Extensions []*extension.Extension
}

// Open finds the databases in the filesystem and opens them.
func (db *DB) Open() error {
	// Load the city database.
	cityMmdb, err := maxminddb.Open("geolite/GeoLite2-City.mmdb")

	if err != nil {
		return err
	}

	db.cityMmdb = cityMmdb

	// Load the ASN database.
	asnMmdb, err := maxminddb.Open("geolite/GeoLite2-ASN.mmdb")

	if err != nil {
		return err
	}

	db.asnMmdb = asnMmdb

	return nil
}

func (db *DB) GetIPInformation(hostname string) (*types.IPRecord, error) {
	// If you are using strings that may be invalid, check that ip is not nil.
	ip := net.ParseIP(hostname)

	// Create an instance of the IP record.
	rec := &types.IPRecord{}

	// Lookup the IP details from the city database.
	err := db.cityMmdb.Lookup(ip, &rec)

	if err != nil {
		return nil, err
	}

	// Lookup the IP details from the ASN database.
	err = db.asnMmdb.Lookup(ip, &rec)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var addlData []any

	for _, ext := range db.Extensions {
		if !ext.IsLookupExtension() {
			continue
		}

		// Here we query the extension for information on the queried IP address. This data will
		// be added on to the response payload in the end as additional information.
		data, _ := ext.RunIPLookup(ip.String())

		if data != nil {
			addlData = append(addlData, data)
		}
	}

	rec.IPAddress = hostname
	rec.AddlData = addlData

	return rec, nil
}

// GetDomainInformation is the old and fast way of getting DNS records.
func (db *DB) GetDomainInformation(hostname string, dnsServerList []string) ([]*types.IPRecord, error) {
	var records []*types.IPRecord = []*types.IPRecord{}

	// Is this a valid domain name?
	if !govalidator.IsDNSName(hostname) {
		// Make sure the request is valid.
		return records, errors.New("invalid input")
	}

	// Perform a DNS lookup.
	ips, _ := dns.DNSLookup(hostname, dnsServerList)

	for i := 0; i < len(ips); i++ {
		// Get the information on the current IP.
		info, err := db.GetIPInformation(ips[i].String())

		if err != nil {
			continue
		}

		// Append the record to the array.
		records = append(records, info)
	}

	return records, nil
}

// GetDomainInfoFromDNS is the new and slower way of getting DNS records.
func (db *DB) GetDomainInfoFromDNS(hostname string, dnsServerList []string, caller dns.DNSCaller) ([]*types.IPRecord, error) {
	var records []*types.IPRecord = []*types.IPRecord{}

	// Is this a valid domain name?
	if !govalidator.IsDNSName(hostname) {
		// Make sure the request is valid.
		return records, errors.New("invalid input")
	}

	// Perform a DNS lookup.
	ips, _ := caller(hostname, dnsServerList)

	for i := 0; i < len(ips); i++ {
		// Get the information on the current IP.
		info, err := db.GetIPInformation(ips[i].String())

		if err != nil {
			continue
		}

		// Append the record to the array.
		records = append(records, info)
	}

	return records, nil
}
