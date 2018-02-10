ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'github.com/libopenstorage/gossip/vendor')
endif

.PHONY: dep

export GO15VENDOREXPERIMENT=1

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v $(PKGS)

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f $(PKGS)

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t $(PKGS)

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -tags "$(TAGS)" -d -v -t -u -f $(PKGS)

dep:
	curl -s -L https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 -o $(GOPATH)/bin/dep
	chmod +x $(GOPATH)/bin/dep
	$(GOPATH)/bin/dep ensure

all: test

test:
	cd proto && go test -v --timeout 1h
