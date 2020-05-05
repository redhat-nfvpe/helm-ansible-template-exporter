# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif
SHELL := /bin/bash
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=helmExport
BINARY_UNIX=$(BINARY_NAME)_unix
export GO111MODULE=on

validate:
	 @if [[ -z "${role}" && -z "${workspace}" && -z "${helm_chart}" ]]; then \
			echo "Please set env variables, (source env.sh)"; \
			exit 1; \
	  fi

dependency:
		$(Q) curl-sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.23.1


tidy:
		$(Q)go mod tidy -v


lint:
		$(Q)golangci-lint run --verbose

test:
		$(Q)go test -v -race ./...


all: build
build:
		$(GOBUILD) -o $(BINARY_NAME) -v ./*.go

.PHONY: clean
clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
		rm -rf workspace
run:
		$(GOBUILD) -o $(BINARY_NAME) -v ./*.go
		./$(BINARY_NAME)

.PHONY: example
 example: validate
		$(GOBUILD) -o $(BINARY_NAME) -v ./*.go
		 ./$(BINARY_NAME) export ${role} --helm-chart=${helm_chart} --workspace=${workspace} --generateFilters=true



