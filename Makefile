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
	cd core-test; cmake ..
	cd core-test; ctest -E "^(tab_|indent_size_|comments_after_|octothorpe_|escaped_octothorpe_|max_property_|max_section_name_|root_file_|unset_)" --output-on-failure .
