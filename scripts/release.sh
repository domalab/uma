#!/bin/bash

# UMA Automated Release Script
# Creates GitHub releases with consistent version management

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to get current version
get_version() {
    if [[ ! -f "$PROJECT_ROOT/VERSION" ]]; then
        print_status $RED "❌ VERSION file not found"
        exit 1
    fi
    cat "$PROJECT_ROOT/VERSION" | tr -d '\n\r' | xargs
}

# Function to validate prerequisites
validate_prerequisites() {
    print_status $BLUE "🔍 Validating prerequisites..."
    
    # Check if gh CLI is installed
    if ! command -v gh &> /dev/null; then
        print_status $RED "❌ GitHub CLI (gh) is not installed"
        print_status $YELLOW "Install with: brew install gh"
        exit 1
    fi
    
    # Check if authenticated with GitHub
    if ! gh auth status &> /dev/null; then
        print_status $RED "❌ Not authenticated with GitHub"
        print_status $YELLOW "Run: gh auth login"
        exit 1
    fi
    
    # Check if version management is consistent
    if ! "$PROJECT_ROOT/scripts/update-version.sh" verify; then
        print_status $RED "❌ Version inconsistencies detected"
        print_status $YELLOW "Run: make version-sync"
        exit 1
    fi
    
    print_status $GREEN "✅ Prerequisites validated"
}

# Function to build plugin package
build_plugin_package() {
    local version=$1
    
    print_status $BLUE "🔨 Building plugin package..."
    
    # Ensure we're in the project root
    cd "$PROJECT_ROOT"
    
    # Sync versions first
    make version-sync
    
    # Build the plugin package
    cd package
    ./create-plugin-package.sh
    
    # Verify files were created
    if [[ ! -f "uma-$version.txz" ]] || [[ ! -f "uma.plg" ]]; then
        print_status $RED "❌ Plugin package build failed"
        exit 1
    fi
    
    print_status $GREEN "✅ Plugin package built successfully"
}

# Function to create GitHub release
create_github_release() {
    local version=$1
    local tag="v$version"
    
    print_status $BLUE "🚀 Creating GitHub release: $tag"
    
    # Check if release already exists
    if gh release view "$tag" &> /dev/null; then
        print_status $YELLOW "⚠️  Release $tag already exists"
        read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status $BLUE "🗑️  Deleting existing release..."
            gh release delete "$tag" --yes
        else
            print_status $YELLOW "Aborting release creation"
            exit 1
        fi
    fi
    
    # Create release notes
    local release_notes="## UMA (Unraid Management Agent) v$version

### 🎯 Plugin Features

- **Professional Display**: \"Unraid Management Agent\" with proper description
- **Signal Icon**: Appropriate icon for monitoring/API service functionality
- **Update Status**: Proper \"up-to-date\" status display in Unraid interface
- **Version Consistency**: Automated version management across all files

### ⚡ Core Functionality

- **Complete API**: 75+ REST API endpoints for system monitoring
- **Docker Management**: Container monitoring, control, and real-time status
- **VM Control**: Virtual machine monitoring and lifecycle management
- **Storage Monitoring**: Array status, disk health, and SMART data
- **UPS Monitoring**: Enhanced protocol detection and status reporting
- **MCP Support**: Model Context Protocol for AI agent integration
- **WebSocket Streaming**: Real-time data streaming capabilities
- **Performance Optimized**: Intelligent polling and reduced logging

### 📦 Installation

**Via Unraid Plugin Manager:**
1. Go to Settings > Plugin Manager
2. Enter plugin URL: \`https://github.com/domalab/uma/releases/download/$tag/uma.plg\`
3. Click Install

**Manual Installation:**
\`\`\`bash
wget https://github.com/domalab/uma/releases/download/$tag/uma.plg
plugin install uma.plg
\`\`\`

### 🔧 Configuration

- **Settings**: Navigate to Settings > Unraid Management Agent
- **API Access**: http://server-ip:34600/api/v1/
- **Documentation**: Available through web interface

This release includes automated version management ensuring consistency across all components."
    
    # Create the release
    gh release create "$tag" \
        --title "UMA v$version - Unraid Management Agent" \
        --notes "$release_notes" \
        --latest
    
    print_status $GREEN "✅ GitHub release created: $tag"
}

# Function to upload release assets
upload_release_assets() {
    local version=$1
    local tag="v$version"
    
    print_status $BLUE "📤 Uploading release assets..."
    
    cd "$PROJECT_ROOT/package"
    
    # Upload plugin files
    gh release upload "$tag" "uma.plg" "uma-$version.txz" --clobber
    
    print_status $GREEN "✅ Release assets uploaded"
}

# Function to verify release
verify_release() {
    local version=$1
    local tag="v$version"
    
    print_status $BLUE "🔍 Verifying release..."
    
    # Check if release exists and has assets
    local asset_count=$(gh release view "$tag" --json assets --jq '.assets | length')
    
    if [[ $asset_count -lt 2 ]]; then
        print_status $RED "❌ Release verification failed: insufficient assets"
        exit 1
    fi
    
    # Test download URLs
    local plugin_url="https://github.com/domalab/uma/releases/download/$tag/uma.plg"
    local package_url="https://github.com/domalab/uma/releases/download/$tag/uma-$version.txz"
    
    if curl -I "$plugin_url" 2>/dev/null | grep -q "200\|302"; then
        print_status $GREEN "✅ Plugin URL accessible: $plugin_url"
    else
        print_status $RED "❌ Plugin URL not accessible: $plugin_url"
        exit 1
    fi
    
    if curl -I "$package_url" 2>/dev/null | grep -q "200\|302"; then
        print_status $GREEN "✅ Package URL accessible: $package_url"
    else
        print_status $RED "❌ Package URL not accessible: $package_url"
        exit 1
    fi
    
    print_status $GREEN "✅ Release verification completed"
}

# Main function
main() {
    local command=${1:-"create"}
    
    case $command in
        "create")
            print_status $BLUE "🚀 Starting automated release process..."
            
            local version=$(get_version)
            print_status $BLUE "📋 Version: $version"
            
            validate_prerequisites
            build_plugin_package "$version"
            create_github_release "$version"
            upload_release_assets "$version"
            verify_release "$version"
            
            print_status $GREEN "🎉 Release process completed successfully!"
            print_status $BLUE "🔗 Release URL: https://github.com/domalab/uma/releases/tag/v$version"
            ;;
            
        "verify")
            local version=$(get_version)
            verify_release "$version"
            ;;
            
        *)
            print_status $BLUE "UMA Automated Release Script"
            echo ""
            echo "Usage: $0 <command>"
            echo ""
            echo "Commands:"
            echo "  create    Create complete GitHub release with assets"
            echo "  verify    Verify existing release"
            echo ""
            echo "Examples:"
            echo "  $0 create     # Create release for current version"
            echo "  $0 verify     # Verify current release"
            ;;
    esac
}

# Run main function with all arguments
main "$@"
