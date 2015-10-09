TAGS := daemon btrfs_noversion
PKGS := $(shell go list ./... | grep -v 'github.com/libopenstorage/openstorage/vendor')

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

vendor:
	go get -v github.com/tools/godep
	rm -rf Godeps
	rm -rf vendor
	# TODO: when godep fixes downloading all tags, remove the custom package
	# https://github.com/tools/godep/issues/271
	godep save $(PKGS) github.com/docker/docker/pkg/chrootarchive
	rm -rf Godeps

build:
	go build -tags "$(TAGS)" $(BUILDFLAGS) $(PKGS)

install:
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

test:
	go test -tags "$(TAGS)" $(PKGS)

docker-build:
	docker build -t openstorage/osd .

docker-test: docker-build
	docker run \
		--privileged \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-v /var/run/docker.sock:/var/run/docker.sock \
		openstorage/osd \
			make test

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
