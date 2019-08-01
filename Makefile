PROJECT_ROOT_DIR := $(CURDIR)
SRC := $(shell git ls-files *.go)

.PHONY: bin test test-go test-core submodule

test: test-go test-core

submodule:
	git submodule update --init

editorconfig: $(SRC)
	go build ./cmd/editorconfig

test-go:
	go test -v ./...

test-core: editorconfig
	cd $(PROJECT_ROOT_DIR)/core-test; cmake ..
	cd $(PROJECT_ROOT_DIR)/core-test; ctest --output-on-failure .
