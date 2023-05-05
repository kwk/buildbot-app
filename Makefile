# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

include ./help.mk

.PHONY: app
## Starts the github app and hot reloads it upon changes
app:
	reflex -r '\.go$$' -s -- sh -c "go build github.com/kwk/buildbot-app/cmd/buildbot-app && ./buildbot-app"

.PHONY: start-infra
## Starts the buildbot containers (1x master, 3x workers).
## The master is automatically reloaded if its master.cfg file changes.
start-infra:
	$(MAKE) -C infra start

.PHONY: stop-infra
## Stops the buildbot containers (1x master, 3x workers).
stop-infra:
	$(MAKE) -C infra stop

.PHONY: forwarder
## Exposes local port 8081 to the public through ngrok
forwarder:
	ngrok http 8081

.PHONY: docs
## Prepares the webiste and the github readme as well as any diagrams
docs:
	$(MAKE) -C docs docs