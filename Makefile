SHELL=/bin/bash
BIN:=./bin
GOLANGCI_LINT_VERSION?=1.24.0

ifeq ($(OS),Windows_NT)
    OSNAME = windows
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        OSNAME = linux
		GOLANGCI_LINT_ARCHIVE=golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz
    endif
    ifeq ($(UNAME_S),Darwin)
        OSNAME = darwin
		GOLANGCI_LINT_ARCHIVE=golangci-lint-$(GOLANGCI_LINT_VERSION)-darwin-amd64.tar.gz
    endif
endif

ifdef os
  OSNAME=$(os)
endif

.PHONY: all
all: lint build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(OSNAME) go build -mod=vendor -ldflags="-s -w" -a -o ./artifacts/psql_admin_utils-unpacked ./cmd/psql_admin_utils
	if [[ "$(OSNAME)" == "linux" ]]; then \
		rm -rf ./artifacts/psql_admin_utils; \
		upx -q -o ./artifacts/psql_admin_utils ./artifacts/psql_admin_utils-unpacked; \
	else \
		mv ./artifacts/psql_admin_utils-unpacked ./artifacts/psql_admin_utils; \
	fi

.PHONY: deps
deps:
	go mod vendor

.PHONY: lint
lint: $(BIN)/golangci-lint/golangci-lint ## lint
	$(BIN)/golangci-lint/golangci-lint run

$(BIN)/golangci-lint/golangci-lint:
	curl -OL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/$(GOLANGCI_LINT_ARCHIVE)
	mkdir -p $(BIN)/golangci-lint/
	tar -xf $(GOLANGCI_LINT_ARCHIVE) --strip-components=1 -C $(BIN)/golangci-lint/
	chmod +x $(BIN)/golangci-lint
	rm -f $(GOLANGCI_LINT_ARCHIVE)
