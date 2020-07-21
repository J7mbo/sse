SHELL := /bin/bash

#
# Public API.
#
build: ## Build the docker container with newly updated code
	@COMPOSE_IGNORE_ORPHANS=True docker-compose --env-file ./infrastructure/dev.env -f ./infrastructure/docker/docker-compose.dev.yml build

mock: ## Generate mocks using go generate and mockery
	@go generate ./internal/

.PHONY: test
test: ## Run unit tests
	@go test ./... -count=1 -v

up: .certs ## Run the docker containers in the background
	@COMPOSE_IGNORE_ORPHANS=True docker-compose --env-file ./infrastructure/dev.env -f ./infrastructure/docker/docker-compose.dev.yml up -d

stop: ## Stop the docker containers
	@COMPOSE_IGNORE_ORPHANS=True docker-compose --env-file ./infrastructure/dev.env -f ./infrastructure/docker/docker-compose.dev.yml stop

kill: ## Kill the docker containers
	@COMPOSE_IGNORE_ORPHANS=True docker-compose --env-file ./infrastructure/dev.env -f ./infrastructure/docker/docker-compose.dev.yml kill

help:
	@echo 'Usage: make [target] ...'
	@echo
	@echo 'targets:'
	@echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"
#
# Private API.
#
.certs:
	@ansible-playbook ./infrastructure/ansible/playbook.yml

.DEFAULT_GOAL := help