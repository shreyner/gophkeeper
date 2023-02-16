PROJECT_DIR = $(shell pwd)
PROJECT_BIN = $(PROJECT_DIR)/bin
$(shell [ -f bin ] || mkdir -p $(PROJECT_BIN))
PATH := $(PROJECT_BIN):$(PATH)

GOLANGCI_LINT = $(PROJECT_BIN)/golangci-lint
TERN = $(PROJECT_BIN)/tern
SWAG = $(PROJECT_BIN)/swag


.PHONY: .install-linter
.install-linter:
	### INSTALL GOLANGCI-LINT ###
	[ -f $(GOLANGCI_LINT) ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_BIN) v1.50.0

.PHONY: lint
lint: .install-linter
	### RUN GOLANGCI-LINT ###
	$(GOLANGCI_LINT) run ./...

.PHONY: .install-migrate
.install-migrate:
	[ -f $(TERN) ] || GOBIN=$(PROJECT_BIN) go install github.com/jackc/tern@latest

.PHONY: migrate
migrate: .install-migrate
	$(TERN) migrate -m migrate

.PHONY: .install-swag
.install-swag:
	### Install swag tool
	[ -f $(SWAG) ] || GOBIN=$(PROJECT_BIN) go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: install-tools
install-tools: .install-linter .install-migrate .install-swag

.PHONY: .install-proto
.install-proto:
	### INSTALL GOLANGCI-LINT ###
	[ -f $(PROJECT_BIN)/protoc-gen-go-grpc ] || GOBIN=$(PROJECT_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	[ -f $(PROJECT_BIN)/protoc-gen-go ] || GOBIN=$(PROJECT_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

.PHONY: proto
proto: .install-proto
	$(PROJECT_BIN)/protoc-gen-go --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/gophkeeper.proto

.PHONY: .install-mockgen
.install-mockgen:
	[ -f $(PROJECT_BIN)/mockgen ] || GOBIN=$(pwd)/bin go install github.com/golang/mock/mockgen@v1.6.0

.PHONY: cert
cert:
	cd ./cert && ./generate_tls.sh

