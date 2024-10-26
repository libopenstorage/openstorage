ifndef PKGS
PKGS := $(shell go list ./... 2>&1 | grep -v 'github.com/libopenstorage/gossip/vendor')
endif

.PHONY: dep

tidy:
	go mod tidy

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
