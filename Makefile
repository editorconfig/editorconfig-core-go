PROJECT_ROOT_DIR := $(CURDIR)
SRC := $(shell git ls-files *.go */*.go)

.PHONY: bin test test-go test-core submodule

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
		-E "^(comments_after_section|(escaped_)?octothorpe_(in_|comments_).*|root_file_mixed_case)$$" \
		--output-on-failure \
		.
