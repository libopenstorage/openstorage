ifeq ($(BUILD_TYPE),debug)
BUILD_OPTIONS= -gcflags "-N -l"
endif

.PHONY: clean all

all: test

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v ./...

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f ./...

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t ./...

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./...

restore:
	godep restore ./...

update-vendor-all:
	go get -u github.com/tools/godep
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./...
	rm -rf Godeps
	rm -rf vendor
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 godep save ./...

build: restore
	go build ./...

lint: restore
	go get -v github.com/golang/lint/golint
	golint ./...

vet: restore
	go vet ./...

errcheck: restore
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

test: restore pretest
	go test ./...

docker-build:
	docker build -t openstorage/osd .

docker-test: docker-build
	docker run openstorage/osd make test

clean:
	go clean ./...

.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	restore \
	update-vendor-all \
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
