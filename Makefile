# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif


GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=helmExport
BINARY_UNIX=$(BINARY_NAME)_unix
HELM_EXAMPLE=./examples/helmcharts/nginx/
ROLENAME=nginx
WORKSPACE=workspace

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
clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
		rm -rf workspace
run:
		$(GOBUILD) -o $(BINARY_NAME) -v ./*.go
		./$(BINARY_NAME)



.PHONY: example
example:
		$(GOBUILD) -o $(BINARY_NAME) -v ./*.go
		./$(BINARY_NAME) export $(ROLENAME) --helm-chart=$(HELM_EXAMPLE) --workspace=$(WORKSPACE) --generateFilters=true