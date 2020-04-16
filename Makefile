GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=helmExport
BINARY_UNIX=$(BINARY_NAME)_unix
HELM_EXAMPLE=./examples/helmcharts/nginx/
ROLENAME=ngnix
WORKSPACE=workspace
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
		./$(BINARY_NAME) export $(ROLENAME) --helm-chart=$(HELM_EXAMPLE) --workspace=$(WORKSPACE)