# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

include ../help.mk

.PHONY: media
media:
	ln -sf ~/bin/plantuml.jar .
	java -jar plantuml.jar -progress -tsvg media/

.PHONY: website
website:
	asciidoctor -r asciidoctor-lists -a docinfo2 README.in.adoc --doctype article -o index.html.tmp
	echo "<!-- DO NOT MODIFY THIS DOCUMENT DIRECTLY! -->" > index.html
	echo "<!-- MODIFY docs/README.in.adoc INSTEAD! --> " >> index.html
	echo "<!-- THEN RUN make docs! -->" >> index.html
	cat index.html.tmp >> index.html
	rm index.html.tmp

.PHONY: github-readme
## This generates the README.adoc to be shown on GitHub 
github-readme:
	asciidoctor -r asciidoctor-lists README.in.adoc --doctype article --backend docbook -o README.xml
	pandoc --from=docbook --to=asciidoc -o README.adoc.tmp README.xml
	echo "// DO NOT MODIFY THIS DOCUMENT DIRECTLY!" > ../README.adoc
	echo "// MODIFY docs/README.in.adoc INSTEAD!" >> ../README.adoc
	echo "// THEN RUN make docs!" >> ../README.adoc
	cat preamble.adoc >> ../README.adoc
	cat README.adoc.tmp >> ../README.adoc
	rm README.adoc.tmp README.xml

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
docs: media website github-readme
	