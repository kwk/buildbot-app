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
## This step might look counter intuitive at first but it has a reason.
## I make heavy use of a the asciidoc include directive to keep the document
## as close to the truth as possible.
## Sometimes I include complete files and sometimes just tagged regions or
## in the worst case, just lines by line number. Unfortunately github doesn't
## allow includes at all. That's why I convert my asciidoc document
## to docbook only to convert it back to asciidoc but this time with
## materialized include files.
##
## See:
##   * https://docs.asciidoctor.org/asciidoc/latest/directives/include/.
##   * https://docs.asciidoctor.org/asciidoc/latest/directives/include-tagged-regions/
##   * https://docs.asciidoctor.org/asciidoc/latest/directives/include-lines/
## 
## Files and their purpose:
##   * index.html - rendered on https://kwk.github.io/buildbot-app/
##   * README.adoc - README on https://github.com/kwk/buildbot-app#readme
docs:
	asciidoctor README.in.adoc --doctype article -o index.html
	# Prepare asciidoc to be rendered on github
	asciidoctor README.in.adoc --doctype article --backend docbook -o README.xml
	pandoc --from=docbook --to=asciidoc -o README.adoc.tmp README.xml
	cat preamble.adoc > README.adoc
	cat README.adoc.tmp >> README.adoc
	rm README.adoc.tmp README.xml