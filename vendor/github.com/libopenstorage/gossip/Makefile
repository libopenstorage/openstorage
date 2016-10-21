ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'github.com/libopenstorage/gossip/vendor')
endif

export GO15VENDOREXPERIMENT=1

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v $(PKGS)

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f $(PKGS)

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t $(PKGS)

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -tags "$(TAGS)" -d -v -t -u -f $(PKGS)

vendor-update:
	GO15VENDOREXPERIMENT=0 GOOS=linux GOARCH=amd64 go get -d -v -t -u -f $(shell go list ./... 2>&1 | grep -v 'github.com/portworx/arturo/vendor')

vendor-without-update:
	go get -v github.com/kardianos/govendor
	rm -rf vendor
	govendor init
	GOOS=linux GOARCH=amd64 govendor add +external
	GOOS=linux GOARCH=amd64 govendor update +vendor
	GOOS=linux GOARCH=amd64 govendor add +external
	GOOS=linux GOARCH=amd64 govendor update +vendor

vendor: vendor-update vendor-without-update

all: test

test:
	ifconfig lo:2 127.0.0.2 netmask 255.255.255.0 up
	ifconfig lo:3 127.0.0.3 netmask 255.255.255.0 up
	ifconfig lo:4 127.0.0.4 netmask 255.255.255.0 up
	ifconfig lo:5 127.0.0.5 netmask 255.255.255.0 up
	ifconfig lo:6 127.0.0.6 netmask 255.255.255.0 up

	cd proto && go test

	ifconfig lo:2 down
	ifconfig lo:3 down
	ifconfig lo:4 down
	ifconfig lo:5 down
	ifconfig lo:6 down
