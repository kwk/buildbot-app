# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

COMPOSE_ARGS := -p "infra" \
				--podman-build-args=--quiet \
				--podman-start-args=--quiet \
				--podman-run-args=--quiet \
				--podman-rm-args=-f

include ../help.mk

.PHONY: build
## Builds all container images we need for local testing
build:
	podman-compose $(COMPOSE_ARGS) build

.PHONY: start
## Spawns a buildbot master and some workerson localhost using podman-compose.
## Upon launch, the buildbot website is opened in your default browser
## and the compose tool's logs are followed in the terminal.
start: build stop
	podman-compose $(COMPOSE_ARGS) up --remove-orphans -d
	@echo -n "Waiting until Buildbot master Web-UI is up."
	@timeout 22 sh -c 'until curl --insecure -Sf https://localhost:8443/#/workers > /dev/null 2>&1; do echo -n "."; sleep 0.1; done'
	@echo "DONE"
	@xdg-open https://localhost:8443/#/workers
	@echo "Opening Buildbot Workers Web UI (https://localhost:8443/#/workers)"
	@echo "Following compose logs"
	podman-compose logs -f

.PHONY: follow-logs
## Follow the logs of the containers
follow-logs:
	@echo "Following compose logs"
	podman-compose logs -f

.PHONY: stop
## Stops and removes any running container images from this job
stop:
	-podman-compose $(COMPOSE_ARGS) down -t0
