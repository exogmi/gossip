# Makefile for Gossip IRC Server

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=gossip-server

# Python parameters
PYTHON=python

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) cmd/server/main.go

test:
	$(GOTEST) ./...

integration-test:
	$(PYTHON) -m unittest tests/integration/test_nickname_change.py

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: all build test integration-test clean
