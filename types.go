package main

type IPRecord struct {
	Country struct {
		ISOCode   string            `maxminddb:"iso_code" json:"iso_code"`
		GeonameID int               `maxminddb:"geoname_id" json:"geoname_id"`
		Name      map[string]string `maxminddb:"names" json:"name"`
	} `maxminddb:"country" json:"country"`
	City struct {
		GeonameID int               `maxminddb:"geoname_id" json:"geoname_id"`
		Name      map[string]string `maxminddb:"names" json:"name"`
	} `maxminddb:"city" json:"city"`
	Location struct {
		Latitude  float32 `maxminddb:"latitude" json:"latitude"`
		Longitude float32 `maxminddb:"longitude" json:"longitude"`
		MetroCode int     `maxminddb:"metro_code" json:"metro_code"`
	} `maxminddb:"location" json:"location"`
	ASN       int    `maxminddb:"autonomous_system_number" json:"asn"`
	Org       string `maxminddb:"autonomous_system_organization" json:"org"`
	IPAddress string `json:"ip_address"`
}

type ApiResponse struct {
	Success bool      `json:"success"`
	Status  string    `json:"status"`
	Record  *IPRecord `json:"record"`
}

type MultiApiResponse struct {
	Success bool        `json:"success"`
	Status  string      `json:"status"`
	Records []*IPRecord `json:"records"`
}

type DNSApiResponse struct {
	Success bool     `json:"success"`
	Servers []string `json:"servers"`
}
