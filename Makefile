ifeq ($(BUILD_TYPE),debug)
BUILD_OPTIONS= -gcflags "-N -l"
else
BUILD_OPTIONS= 
endif
.PHONY: clean all
TARGETS := openstorage

export BASE_DIR=$(shell git rev-parse --show-toplevel)
export GOPATH=$(shell echo ${BASE_DIR}| sed 's@\(.*\)/src/github.com.*@\1@g')

all: $(TARGETS) tags

tags:
	@ctags -R 

openstorage:
	@echo "Building openstorage..."
	@echo go build $(BUILD_OPTIONS) -tags daemon  -o osd 
	@go build $(BUILD_OPTIONS) -tags daemon  -o osd 

docker:
	@docker rmi -f osd || true
	@docker build -t osd -f Dockerfile .

test:
	@go test ./... -tags daemon

clean:
	@echo "Cleaning openstorage..."
	@rm -f tags
	@rm -f osd
