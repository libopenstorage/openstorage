ifeq ($(BUILD_TYPE),debug)
BUILD_OPTIONS= -compiler gccgo --gccgoflags -static-libgo
else
BUILD_OPTIONS= 
endif
.PHONY: clean all
TARGETS := libopenstorage

all: $(TARGETS) tags

tags:
	@ctags -R 

libopenstorage:
	@echo "Building libopenstorage..."
	@go build $(BUILD_OPTIONS)  -o libopenstorage

clean:
	@echo "Cleaning libopenstorage..."
	@rm -f tags
	@rm -f libopenstorage
