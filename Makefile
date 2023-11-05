#!/usr/bin/make -f
SHELL = /bin/sh

#### Start of system configuration section. ####

srcdir = .

GO ?= go
GOFLAGS ?= -buildvcs=true
EXECUTABLE ?= hub

#### End of system configuration section. ####

.PHONY: all
all: main.go
	$(GO) build -v $(GOFLAGS) -o $(EXECUTABLE)

.PHONY: clean
clean: ## Delete all files in the current directory that are normally created by building the program
	$(GO) clean

.PHONY: check
check: ## Perform self-tests
	$(GO) test -v -cover -failfast -short -shuffle=on $(GOFLAGS) $(srcdir)/...

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
