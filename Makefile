all:
	@ go build .

clean:
	@ rm geoip-service

prep:
	go get github.com/oschwald/maxminddb-golang
	go get github.com/gorilla/mux
	go get github.com/asaskevich/govalidator

