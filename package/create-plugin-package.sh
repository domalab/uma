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
SRC_DIR="$SCRIPT_DIR/../src"

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

    # Check if modern src directory exists
    if [ ! -d "$SRC_DIR" ]; then
        log_error "Modern src/ directory not found at $SRC_DIR"
        log_error "Please ensure the modern plugin structure has been created"
        exit 1
    fi

    # Check if UMA binary exists in src structure
    if [ ! -f "$SRC_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/uma" ]; then
        log_error "UMA binary not found in modern structure"
        log_error "Please build the binary first: make release"
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
    log_info "Copying plugin files from modern src/ structure..."

    # Copy from modern src/ directory structure
    log_info "Using modern src/ directory structure"
    cp -r "$SRC_DIR/"* "$BUILD_DIR/"
    
    # Set proper permissions
    chmod +x "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/uma"
    chmod +x "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/scripts/"*
    chmod +x "$BUILD_DIR/usr/local/emhttp/plugins/$PLUGIN_NAME/event/"*
    
    # Create slack-desc file for package (only if not using modern src/ structure)
    if [ ! -f "$BUILD_DIR/install/slack-desc" ]; then
        cat > "$BUILD_DIR/install/slack-desc" << EOF
        |-----handy-ruler------------------------------------------------------|
uma: uma (Unraid Management Agent)
uma:
uma: System monitoring and management API for Unraid servers
uma:
uma: Features:
uma: - 75+ REST API endpoints for system monitoring
uma: - Docker container and VM management
uma: - Storage array and UPS monitoring
uma: - Model Context Protocol (MCP) support
uma: - Real-time WebSocket streaming
uma:
EOF
    fi
    
    log_success "Plugin files copied"
}

# Function to create package archive
create_package() {
    log_info "Creating package archive..."

    # Modern Unraid package naming convention
    local modern_package="$SCRIPT_DIR/${PLUGIN_NAME}-${VERSION}-noarch-1.txz"
    local legacy_package="$SCRIPT_DIR/${PLUGIN_NAME}-${VERSION}.txz"

    # Create the package with modern standards
    cd "$BUILD_DIR"
    tar --owner=0 --group=0 -cJf "$modern_package" .

    # Also create legacy format for compatibility
    tar --owner=0 --group=0 -czf "$legacy_package" .
    cd "$SCRIPT_DIR"

    # Generate SHA256 checksum (modern security standard)
    local sha256_hash
    sha256_hash=$(sha256sum "$modern_package" | awk '{print $1}')

    # Generate MD5 for legacy compatibility
    local md5_hash
    md5_hash=$(md5sum "$legacy_package" | awk '{print $1}')

    log_success "Package created: $(basename "$modern_package")"
    log_success "Legacy package created: $(basename "$legacy_package")"
    log_info "SHA256 checksum: $sha256_hash"
    log_info "MD5 checksum: $md5_hash"

    # Update .plg file with correct MD5 (for legacy compatibility)
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
