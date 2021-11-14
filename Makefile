# 
# Makefile for hardware-events
# 
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
GOPATH?=`$(GOCMD) env GOPATH`

CONFIG=config-local.yaml
TEMPLATES=*.go.txt
BINARY=hardware-events
TESTS=./...
COVERAGE_FILE=coverage.out
DEPLOY=~/hardware-events

BUILD_DATE=`date`
BUILD_COMMIT=`git rev-parse HEAD`

.PHONY: all test build coverage clean deploy

all: test build

build:
		$(GOBUILD) -race -o $(BINARY) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

test:
		$(GOTEST) -v -race $(TESTS)

coverage:
		$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
		$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
		$(GOCLEAN)
		rm -rf $(BINARY) $(COVERAGE_FILE) dist/*

deploy: $(CONFIG) build
		@mkdir -p $(DEPLOY)
		cp -av $(BINARY) $(CONFIG) $(TEMPLATES) hardware-events.service $(DEPLOY)/

deploy-%: build config-%.yaml
		@echo Deploying $*...
		rsync -avz $(BINARY) config-$*.yaml $(TEMPLATES) hardware-events.service $*:$(DEPLOY)/
