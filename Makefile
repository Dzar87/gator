SHELL:=/bin/bash

# Directory where this file resides. Allows running the Makefile from anywhere using
# the `-f` parameter
ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
DB_USER ?= gator
DB_PASS ?= gator_pass
DB_NAME ?= gator
DB_PORT ?= 5432
GATOR_COMPOSE_CMD ?= docker compose

testfile ?= "./..."

all:
	@make help

.PHONY: help
help:  ## Show this help message
	@echo "  Makefile targets for:"
	@echo "  File: $(ROOT_DIR)/Makefile"
	@echo
	@make help-targets

.PHONY: help-targets
help-targets:
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test:
	go test $(testfile)

.PHONY: test-cov
test-cov:
	go test -coverprofile=.cover $(testfile) \
	    && go tool cover -func=.cover \
		&& unlink .cover

.PHONY: vet
vet:
	go vet $(testfile)

.PHONY: fmt
fmt:
	test -z "$(shell go fmt . ./internal/...)"

.PHONY: build
build:
	go build . ./internal/...

.PHONY: up
up:  ## Run compose up in detached mode
	$(GATOR_COMPOSE_CMD) up -d

.PHONY: down
down:  ## Run compose down
	$(GATOR_COMPOSE_CMD) down

# POSTGRES
.PHONY: psql
psql:  ## Run psql on the postgres container
	$(GATOR_COMPOSE_CMD) exec -it postgres psql -U $(DB_USER) -d postgres $(args)

.PHONY: psql-notty
psql-notty:  ## Run psql on the postgres container (no-tty)
	$(GATOR_COMPOSE_CMD) exec -T postgres psql -U $(DB_USER) $(args)

.PHONY: psql-url
psql-url:  ## Print the Postgres connection string
	@echo "postgres://$(DB_USER):$(DB_PASS)@localhost:$(DB_PORT)/$(DB_NAME)"
