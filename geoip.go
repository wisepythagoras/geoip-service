package main

import (
	"github.com/oschwald/maxminddb-golang"
)

// OpenCityDB: Open the city database.
func OpenCityDB() (*maxminddb.Reader, error) {
	// Load the database.
	cityMmdb, err := maxminddb.Open("geolite/GeoLite2-City.mmdb")

	if err != nil {
		return nil, err
	}

	return cityMmdb, nil
}

// OpenASNDB: Open the ASN database.
func OpenASNDB() (*maxminddb.Reader, error) {
	// Load the database.
	asnMmdb, err := maxminddb.Open("geolite/GeoLite2-ASN.mmdb")

	if err != nil {
		return nil, err
	}

	return asnMmdb, nil
}
