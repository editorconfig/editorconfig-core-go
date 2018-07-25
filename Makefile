PROJECT_ROOT_DIR := $(CURDIR)
SRC := editorconfig.go cmd/editorconfig/main.go

.PHONY: bin check test test-core submodule

check: test test-core

submodule:
	git submodule update --init

editorconfig: $(SRC)
	go build ./cmd/editorconfig

test:
	go test -v

test-core: editorconfig
	cd $(PROJECT_ROOT_DIR)/core-test && \
		cmake -DEDITORCONFIG_CMD="$(PROJECT_ROOT_DIR)/editorconfig" .
# Temporarily disable core-test
	# cd $(PROJECT_ROOT_DIR)/core-test && \
	# 	ctest --output-on-failure .
