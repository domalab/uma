#!/bin/bash

# UMA Build Script with Version Management
# This script provides a single source of truth for version management

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Version management
VERSION_FILE="$PROJECT_ROOT/VERSION"
PLG_FILE="$PROJECT_ROOT/uma.plg"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Read version from VERSION file
get_version() {
    if [[ ! -f "$VERSION_FILE" ]]; then
        log_error "VERSION file not found at $VERSION_FILE"
        exit 1
    fi
    
    local version=$(cat "$VERSION_FILE" | tr -d '\n\r' | xargs)
    if [[ -z "$version" ]]; then
        log_error "VERSION file is empty"
        exit 1
    fi
    
    echo "$version"
}

# Validate version format (YYYY.MM.DD)
validate_version() {
    local version="$1"
    if [[ ! "$version" =~ ^[0-9]{4}\.[0-9]{2}\.[0-9]{2}$ ]]; then
        log_error "Invalid version format: $version (expected YYYY.MM.DD)"
        exit 1
    fi
}

# Update PLG file with current version
update_plg_version() {
    local version="$1"
    log_info "Updating PLG file with version $version"
    
    # Update version entity in PLG file
    sed -i.bak "s/<!ENTITY version[[:space:]]*\"[^\"]*\">/<!ENTITY version   \"$version\">/" "$PLG_FILE"
    
    # Update version in post-install script
    sed -i.bak "s/Version: [0-9]\{4\}\.[0-9]\{2\}\.[0-9]\{2\}/Version: $version/" "$PLG_FILE"
    
    # Remove backup file
    rm -f "$PLG_FILE.bak"
    
    log_success "PLG file updated with version $version"
}

# Build binary with version injection
build_binary() {
    local version="$1"
    local target="${2:-uma}"
    
    log_info "Building $target with version $version"
    
    cd "$PROJECT_ROOT"
    
    # Build with version injection
    go build -ldflags "-X main.Version=$version" -o "$target"
    
    if [[ $? -eq 0 ]]; then
        log_success "Binary built successfully: $target"
        log_info "Version: $version"
    else
        log_error "Build failed"
        exit 1
    fi
}

# Build for Linux (Unraid target)
build_linux() {
    local version="$1"
    local target="${2:-uma}"
    
    log_info "Building Linux binary $target with version $version"
    
    cd "$PROJECT_ROOT"
    
    # Build for Linux AMD64
    GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$version" -o "$target"
    
    if [[ $? -eq 0 ]]; then
        log_success "Linux binary built successfully: $target"
        log_info "Version: $version"
    else
        log_error "Linux build failed"
        exit 1
    fi
}

# Main build function
main() {
    local command="${1:-build}"
    local target="${2:-uma}"
    
    log_info "UMA Build Script - Single Source of Truth Version Management"
    log_info "Project root: $PROJECT_ROOT"
    
    # Get and validate version
    local version=$(get_version)
    validate_version "$version"
    
    log_info "Current version: $version"
    
    case "$command" in
        "build")
            update_plg_version "$version"
            build_binary "$version" "$target"
            ;;
        "linux")
            update_plg_version "$version"
            build_linux "$version" "$target"
            ;;
        "version")
            echo "$version"
            ;;
        "update-plg")
            update_plg_version "$version"
            ;;
        *)
            echo "Usage: $0 [build|linux|version|update-plg] [target]"
            echo ""
            echo "Commands:"
            echo "  build      - Build binary for current platform (default)"
            echo "  linux      - Build binary for Linux (Unraid target)"
            echo "  version    - Display current version"
            echo "  update-plg - Update PLG file with current version"
            echo ""
            echo "Arguments:"
            echo "  target     - Output binary name (default: uma)"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
