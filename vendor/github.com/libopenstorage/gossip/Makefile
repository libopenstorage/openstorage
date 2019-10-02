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
	for pkg in $(PKGS); \
	do \
		go test --timeout 1h -v -tags unittest -coverprofile=profile.out -covermode=atomic $(BUILD_OPTIONS) $${pkg} || exit 1; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done

docker-test:
	docker run \
	--privileged \
        --net=host \
	openstorage/osd-gossip \
		make test

docker-build-osd-gossip:
	docker build -t openstorage/osd-gossip -f Dockerfile.osd-gossip .

push-docker-images: docker-build-osd-gossip
	docker push openstorage/osd-gossip
