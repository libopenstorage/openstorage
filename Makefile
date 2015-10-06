ifeq ($(BUILD_TYPE),debug)
BUILD_OPTIONS= -gcflags "-N -l"
endif

all: test

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v ./...

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f ./...

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t ./...

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./...

vendor:
	go get -u github.com/tools/godep
	-CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./...
	rm -rf Godeps
	rm -rf vendor
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 godep save \
							./... \
							github.com/docker/docker/pkg/chrootarchive

build:
	go build -tags daemon $(shell go list ./... | grep -v 'openstorage/vendor')

lint:
	go get -v github.com/golang/lint/golint
	golint $(shell go list ./... | grep -v 'openstorage/vendor')

vet:
	go vet $(shell go list ./... | grep -v 'openstorage/vendor')

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck $(shell go list ./... | grep -v 'openstorage/vendor')

pretest: lint vet errcheck

test:
	go test -tags daemon $(shell go list ./... | grep -v 'openstorage/vendor')

docker-build:
	docker build -t openstorage/osd .

docker-test: docker-build
	docker run --privileged openstorage/osd make test

clean:
	go clean $(shell go list ./... | grep -v 'openstorage/vendor')

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
