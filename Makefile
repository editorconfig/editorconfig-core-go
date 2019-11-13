PROJECT_ROOT_DIR := $(CURDIR)
SRC := $(shell git ls-files *.go */*.go)

.PHONY: bin test test-go test-core test-skipped submodule

test: test-go test-core

submodule:
	git submodule update --init

editorconfig: $(SRC)
	go build ./cmd/editorconfig

test-go:
	go test -v ./...

test-core: editorconfig
	cd core-test; \
		cmake ..
	cd core-test; \
		ctest \
		-E "^(comments_after_section|(escaped_)?octothorpe_(in_|comments_).*)$$" \
		--output-on-failure \
		.

test-skipped: editorconfig
	cd core-test; \
		ctest \
		-R "^(comments_after_section|(escaped_)?octothorpe_(in_|comments_).*)$$" \
		--show-only \
		.
