#!/bin/bash

# In your .bashrc or .bash_profile add your license key.
# export GEOLITE_LICENSE_KEY="YOUR LICENSE KEY"
URL="https://download.maxmind.com/app/geoip_download"

wget "$URL?edition_id=GeoLite2-ASN&license_key=$GEOLITE_LICENSE_KEY&suffix=tar.gz" -O GeoLite2-ASN_$(date +%Y%m%d).tar.gz
wget "$URL?edition_id=GeoLite2-City&license_key=$GEOLITE_LICENSE_KEY&suffix=tar.gz" -O GeoLite2-City_$(date +%Y%m%d).tar.gz
wget "$URL?edition_id=GeoLite2-Country&license_key=$GEOLITE_LICENSE_KEY&suffix=tar.gz" -O GeoLite2-Country_$(date +%Y%m%d).tar.gz
