#!/bin/bash

tar -xvf GeoLite2-City_$(date +%Y)*.tar.gz
tar -xvf GeoLite2-Country_$(date +%Y)*.tar.gz
tar -xvf GeoLite2-ASN_$(date +%Y)*.tar.gz

cp GeoLite2-*_$(date +%Y)*/*.mmdb geolite/

rm -rf GeoLite2-*
