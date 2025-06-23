#
# Makefile to perform "live code reloading" after changes to .go files.
#
# To start live reloading run the following command:
# $ make serve
#

mb_date := $(shell date '+%Y.%m.%d')
mb_hash := $(shell git rev-parse --short HEAD)

# binary name to kill/restart
PROG = uma

# targets not associated with files
.PHONY: default build test coverage clean kill restart serve validate-schemas build-validator install-schema-tools version-sync version-set version-verify

# default targets to run when only running `make`
default: test

# clean up
clean:
	go clean

local: clean version-sync
	go build fmt
	go build -ldflags "-X main.Version=$(shell cat VERSION)" -v -o ${PROG}

release: clean version-sync
	go build fmt
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(shell cat VERSION)" -v -o ${PROG}

# run unit tests with code coverage
test:
	go test -v

# generate code coverage report
coverage: test
	go build test -coverprofile=.coverage.out
	go build tool cover -html=.coverage.out

# attempt to kill running server
kill:
	-@killall -9 $(PROG) 2>/dev/null || true

# attempt to build and start server
restart:
	@make kill
	@make build; (if [ "$$?" -eq 0 ]; then (env GIN_MODE=debug ./${PROG} &); fi)

publish: build
	cp ./${PROG} ~/bin

# Version management targets
version-sync:
	@echo "Synchronizing version across all files..."
	@./scripts/update-version.sh sync

version-set:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make version-set VERSION=2025.06.24"; \
		exit 1; \
	fi
	@echo "Setting new version: $(VERSION)"
	@./scripts/update-version.sh set $(VERSION)

version-verify:
	@echo "Verifying version consistency..."
	@./scripts/update-version.sh verify

version-current:
	@./scripts/update-version.sh current

version-help:
	@echo "UMA Version Management System"
	@echo ""
	@echo "Commands:"
	@echo "  make version-current          Display current version"
	@echo "  make version-set VERSION=X    Set new version and sync all files"
	@echo "  make version-sync             Sync current version across all files"
	@echo "  make version-verify           Verify version consistency"
	@echo "  make version-help             Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make version-set VERSION=2025.06.24    # Set new version"
	@echo "  make version-sync                       # Sync existing version"
	@echo "  make version-verify                     # Check consistency"
	@echo ""
	@echo "Release Management:"
	@echo "  ./scripts/release.sh create             # Create GitHub release"
	@echo "  ./scripts/release.sh verify             # Verify release"
	@echo ""
	@echo "Documentation: docs/version-management.md"

# Schema validation targets
install-schema-tools:
	@echo "Installing OpenAPI validation tools..."
	@echo "Installing modern OpenAPI CLI tools (zero deprecated dependencies)..."
	npm install -g @redocly/cli @pb33f/openapi-changes
	@echo "✅ Installed @redocly/cli (provides 'openapi' and 'redocly' commands)"
	@echo "✅ Installed @pb33f/openapi-changes (provides 'openapi-changes' command)"
	@echo ""
	@echo "Available commands:"
	@echo "  openapi lint <spec>                    - Validate and lint OpenAPI specification"
	@echo "  redocly lint <spec>                    - Alternative command for validation"
	@echo "  openapi-changes summary <old> <new>   - Compare OpenAPI specifications"

build-validator:
	@echo "Building schema validator..."
	@mkdir -p tools/schema-validator
	cd tools/schema-validator && go build -o schema-validator main.go

validate-schemas: build-validator
	@echo "Running API schema validation..."
	@chmod +x tools/validate-schemas.sh
	./tools/validate-schemas.sh

validate-schemas-remote:
	@echo "Running API schema validation against remote server..."
	@chmod +x tools/validate-schemas.sh
	UMA_HOST=$(UMA_HOST) UMA_PORT=$(UMA_PORT) ./tools/validate-schemas.sh

schema-help:
	@echo "Schema Validation Targets:"
	@echo "  install-schema-tools  - Install modern OpenAPI validation tools (@redocly/cli, @pb33f/openapi-changes)"
	@echo "  build-validator      - Build the schema validator tool"
	@echo "  validate-schemas     - Run schema validation against local UMA"
	@echo "  validate-schemas-remote - Run validation against remote UMA"
	@echo ""
	@echo "Environment Variables:"
	@echo "  UMA_HOST            - Target UMA host (default: 192.168.20.21)"
	@echo "  UMA_PORT            - Target UMA port (default: 34600)"
	@echo ""
	@echo "Examples:"
	@echo "  make install-schema-tools  # Install modern tools (zero deprecated packages)"
	@echo "  make validate-schemas      # Validate against live Unraid server"
	@echo "  UMA_HOST=localhost make validate-schemas-remote  # Validate against localhost"
	@echo ""
	@echo "Installed Tools:"
	@echo "  openapi lint <spec>                    # Validate and lint OpenAPI spec"
	@echo "  redocly lint <spec>                    # Alternative command for validation"
	@echo "  openapi-changes summary <old> <new>   # Compare two OpenAPI specifications"
