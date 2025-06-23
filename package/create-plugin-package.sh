#!/bin/bash
#
# UMA Plugin Package Creator
# Creates a complete Unraid plugin package ready for distribution
#

set -e

# Configuration
PLUGIN_NAME="uma"
AUTHOR="domalab"

# Get version from single source of truth
if [[ -f "../VERSION" ]]; then
    VERSION=$(cat "../VERSION" | tr -d '\n\r' | xargs)
else
    echo "ERROR: VERSION file not found. Run 'make version-sync' first."
    exit 1
fi
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
PACKAGE_DIR="$SCRIPT_DIR/uma"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if we're in the right directory
    if [ ! -f "$PACKAGE_DIR/uma" ]; then
        log_error "UMA binary not found at $PACKAGE_DIR/uma"
        log_error "Please ensure you're running this script from the package directory"
        exit 1
    fi
    
    # Check if required tools are available
    local missing_tools=()
    
    if ! command -v tar >/dev/null 2>&1; then
        missing_tools+=("tar")
    fi
    
    if ! command -v xz >/dev/null 2>&1; then
        missing_tools+=("xz")
    fi
    
    if ! command -v md5sum >/dev/null 2>&1; then
        missing_tools+=("md5sum")
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to prepare build directory
prepare_build_dir() {
    log_info "Preparing build directory..."
    
    # Clean and create build directory
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    # Create package structure
    mkdir -p "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME"
    mkdir -p "$BUILD_DIR/install"
    
    log_success "Build directory prepared"
}

# Function to copy plugin files
copy_plugin_files() {
    log_info "Copying plugin files..."
    
    # Copy all plugin files
    cp -r "$PACKAGE_DIR"/* "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/"
    
    # Set proper permissions
    chmod +x "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/uma"
    chmod +x "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/scripts/"*
    chmod +x "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/event/"*
    
    # Create slack-desc file for package
    cat > "$BUILD_DIR/install/slack-desc" << EOF
# HOW TO EDIT THIS FILE:
# The "handy ruler" below makes it easier to edit a package description.
# Line up the first '|' above the ':' following the base package name, and
# the '|' on the right side marks the last column you can put a character in.
# You must make exactly 11 lines for the formatting to be correct.  It's also
# customary to leave one space after the ':' except on otherwise blank lines.

    |-----handy-ruler------------------------------------------------------|
uma: UMA (Unraid Management Agent)
uma:
uma: Comprehensive system monitoring and management for Unraid servers
uma: through a modern REST API, WebSocket streaming, and MCP protocol.
uma:
uma: Features: System monitoring, Docker management, VM control, UPS
uma: monitoring, storage array monitoring, real-time data streaming,
uma: MCP support for AI agents, optimized performance, and clean web UI.
uma:
uma: Author: $AUTHOR
uma: Version: $VERSION
EOF
    
    log_success "Plugin files copied"
}

# Function to create package archive
create_package() {
    log_info "Creating package archive..."
    
    local package_file="$SCRIPT_DIR/${PLUGIN_NAME}-${VERSION}.txz"
    
    # Create the package
    cd "$BUILD_DIR"
    tar -cJf "$package_file" .
    cd "$SCRIPT_DIR"
    
    # Generate MD5 checksum
    local md5_hash
    md5_hash=$(md5sum "$package_file" | awk '{print $1}')
    
    log_success "Package created: $(basename "$package_file")"
    log_info "MD5 checksum: $md5_hash"
    
    # Update .plg file with correct MD5
    if [ -f "$SCRIPT_DIR/uma.plg" ]; then
        sed -i.bak "s/<!ENTITY md5.*>/<!ENTITY md5       \"$md5_hash\">/" "$SCRIPT_DIR/uma.plg"
        log_success "Updated .plg file with MD5 checksum"
    fi
    
    return 0
}

# Function to validate package
validate_package() {
    log_info "Validating package..."
    
    local package_file="$SCRIPT_DIR/${PLUGIN_NAME}-${VERSION}.txz"
    
    if [ ! -f "$package_file" ]; then
        log_error "Package file not found: $package_file"
        return 1
    fi
    
    # Test package extraction
    local test_dir="$BUILD_DIR/test"
    mkdir -p "$test_dir"
    
    if tar -xJf "$package_file" -C "$test_dir"; then
        log_success "Package extraction test passed"
        
        # Check if key files exist
        local key_files=(
            "usr/local/emhttp/plugins/$PLUGIN_NAME/uma"
            "usr/local/emhttp/plugins/$PLUGIN_NAME/uma.page"
            "usr/local/emhttp/plugins/$PLUGIN_NAME/scripts/start"
            "usr/local/emhttp/plugins/$PLUGIN_NAME/scripts/stop"
            "install/slack-desc"
        )
        
        local missing_files=()
        for file in "${key_files[@]}"; do
            if [ ! -f "$test_dir/$file" ]; then
                missing_files+=("$file")
            fi
        done
        
        if [ ${#missing_files[@]} -gt 0 ]; then
            log_error "Missing files in package: ${missing_files[*]}"
            return 1
        fi
        
        log_success "Package validation passed"
    else
        log_error "Package extraction test failed"
        return 1
    fi
    
    return 0
}

# Function to display summary
display_summary() {
    log_info "Package creation summary:"
    echo
    echo "Plugin Name: $PLUGIN_NAME"
    echo "Version: $VERSION"
    echo "Author: $AUTHOR"
    echo
    echo "Generated files:"
    echo "  - ${PLUGIN_NAME}-${VERSION}.txz (package archive)"
    echo "  - uma.plg (plugin definition file)"
    echo
    echo "Installation instructions:"
    echo "  1. Upload both files to a web server or GitHub release"
    echo "  2. Install via Unraid Plugin Manager using the .plg URL"
    echo "  3. Or install manually: 'plugin install uma.plg'"
    echo
    log_success "UMA plugin package created successfully!"
}

# Main execution
main() {
    echo "=== UMA Plugin Package Creator ==="
    echo "Version: $VERSION"
    echo "Author: $AUTHOR"
    echo
    
    check_prerequisites
    prepare_build_dir
    copy_plugin_files
    create_package
    
    if validate_package; then
        display_summary
        exit 0
    else
        log_error "Package validation failed"
        exit 1
    fi
}

# Execute main function
main "$@"
