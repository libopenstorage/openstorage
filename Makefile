TAGS := daemon btrfs_noversion
PKGS := $(shell go list ./... | grep -v 'github.com/openstorage/openstorage/vendor')

ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS := -gcflags "-N -l"
endif

export GO15VENDOREXPERIMENT=1

all: test install

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v $(PKGS)

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f $(PKGS)

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t $(PKGS)

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f $(PKGS)

vendor: update-test-deps
	go get -v github.com/tools/godep
	rm -rf Godeps
	rm -rf vendor
	godep save $(PKGS)

build: deps
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

install: deps
	go install -tags "$(TAGS)" $(PKGS)

lint:
	go get -v github.com/golang/lint/golint
	golint $(PKGS)

vet:
	go vet $(PKGS)

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck $(PKGS)

pretest: lint vet errcheck

test: test-deps
	go test -tags "$(TAGS)" $(PKGS)

docker-build:
	docker build -t openstorage/osd .

docker-test: docker-build
	docker run openstorage/osd make test

clean:
	go clean -i $(PKGS)

.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	vendor \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	docker-build \
	docker-test \
	clean
