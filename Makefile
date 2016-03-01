ifndef TAGS
TAGS := daemon
endif

ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'github.com/libopenstorage/openstorage/vendor')
endif

ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS := -gcflags "-N -l"
endif

ifdef HAVE_BTRFS
TAGS+=btrfs_noversion have_btrfs
endif

ifdef HAVE_CHAINFS
TAGS+=have_chainfs
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

vendor:
	go get -v github.com/kardianos/govendor
	rm -rf vendor
	govendor init
	GOOS=linux GOARCH=amd64 govendor add +external
	GOOS=linux GOARCH=amd64 govendor update +vendor

build:
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

install:
	go install -tags "$(TAGS)" $(PKGS)

proto:
	go get -v go.pedge.io/protoeasy/cmd/protoeasy
	go get -v go.pedge.io/pkg/cmd/strip-package-comments
	protoeasy --exclude vendor --go --go-import-path github.com/libopenstorage/openstorage --grpc --grpc-gateway .
	find . -name *\.pb\*\.go | xargs strip-package-comments

lint:
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v vendor | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go'); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet:
	go vet $(PKGS)

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck -tags "$(TAGS)" $(PKGS)

pretest: lint vet errcheck

test:
	go test -tags "$(TAGS)" $(TESTFLAGS) $(PKGS)

docker-build-osd-dev:
	docker build -t openstorage/osd-dev -f Dockerfile.osd-dev .

docker-build: docker-build-osd-dev
	docker run \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		openstorage/osd-dev \
			make build

docker-test: docker-build-osd-dev
	docker run \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		-e "TESTFLAGS=$(TESTFLAGS)" \
		openstorage/osd-dev \
			make test

docker-build-osd-internal:
	rm -rf _tmp
	mkdir -p _tmp
	go build -a -tags "$(TAGS)" -o _tmp/osd cmd/osd/main.go
	docker build -t openstorage/osd -f Dockerfile.osd .

docker-build-osd: docker-build-osd-dev
	docker run \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e "TAGS=$(TAGS)" \
		-e "PKGS=$(PKGS)" \
		-e "BUILDFLAGS=$(BUILDFLAGS)" \
		openstorage/osd-dev \
			make docker-build-osd-internal

launch: docker-build-osd
	docker run \
		--privileged \
		-d \
		-v $(shell pwd):/etc \
		-v /run/docker/plugins:/run/docker/plugins \
		-v /var/lib/osd/:/var/lib/osd/\
		openstorage/osd -d -f /etc/config.yaml

launch-local-btrfs: install
	sudo bash -x etc/btrfs/init.sh
	sudo $(shell which osd) -d -f config_btrfs.yaml

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
	proto \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	docker-build-osd-dev \
	docker-build \
	docker-test \
	docker-build-osd-internal \
	docker-build-osd \
	launch \
	launch-local-btrfs \
	clean
