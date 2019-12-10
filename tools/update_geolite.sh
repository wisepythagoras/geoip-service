#!/bin/bash

curl http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz --output GeoLite2-City.tar.gz
curl http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz --output GeoLite2-Country.tar.gz
curl http://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz --output GeoLite2-ASN.tar.gz

tar xf GeoLite2-City.tar.gz
tar xf GeoLite2-Country.tar.gz
tar xf GeoLite2-ASN.tar.gz

cp GeoLite2-*_$(date +%Y)*/*.mmdb geolite/

rm -rf GeoLite2-*
