ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v test)
endif

all: test

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

build: deps
	go build ./...

install: deps
	go install ./...

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck \
		github.com/portworx/kvdb \
		github.com/portworx/kvdb/common \
		github.com/portworx/kvdb/consul \
		github.com/portworx/kvdb/mem

pretest: errcheck lint vet

gotest:
	for pkg in $(PKGS); \
		do \
			go test --timeout 1h -v -tags unittest -coverprofile=profile.out -covermode=atomic $(BUILD_OPTIONS) $${pkg} || exit 1; \
			if [ -f profile.out ]; then \
				cat profile.out >> coverage.txt; \
				rm profile.out; \
			fi; \
		done

test: pretest gotest

docker-build-kvdb-dev:
	docker build -t portworx/kvdb:test_container -f $(GOPATH)/src/github.com/portworx/kvdb/Dockerfile.kvdb .

docker-test:
	docker run \
		-v $(GOPATH)/src/github.com/portworx/kvdb:/go/src/github.com/portworx/kvdb \
		portworx/kvdb:test_container \
		make gotest

clean:
	go clean -i ./...

.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	gotest \
	test \
	clean
