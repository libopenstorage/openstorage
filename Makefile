ifeq ($(BUILD_TYPE),debug)
BUILD_OPTIONS= -compiler gccgo --gccgoflags -static-libgo
else
BUILD_OPTIONS= 
endif
.PHONY: clean all
TARGETS := openstorage

all: $(TARGETS) tags

tags:
	@ctags -R 

openstorage:
	@echo "Building openstorage..."
	@go build $(BUILD_OPTIONS) -tags daemon  -o build/osd 

docker:
	@docker build -t osd -f Dockerfile .

test:
	@go test ./...

clean:
	@echo "Cleaning openstorage..."
	@rm -f tags
	@rm -f osd
