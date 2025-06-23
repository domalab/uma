#!/bin/bash

# UMA Automated Version Management System
# This script ensures version consistency across all project files

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

# Version file (single source of truth)
VERSION_FILE="$PROJECT_ROOT/VERSION"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to validate version format
validate_version_format() {
    local version=$1
    if [[ ! $version =~ ^[0-9]{4}\.[0-9]{2}\.[0-9]{2}$ ]]; then
        print_status $RED "❌ Invalid version format: $version"
        print_status $YELLOW "Expected format: YYYY.MM.DD (e.g., 2025.06.23)"
        exit 1
    fi
}

# Function to read version from source file
get_current_version() {
    if [[ ! -f "$VERSION_FILE" ]]; then
        print_status $RED "❌ Version file not found: $VERSION_FILE"
        exit 1
    fi
    
    local version=$(cat "$VERSION_FILE" | tr -d '\n\r' | xargs)
    validate_version_format "$version"
    echo "$version"
}

# Function to update version in uma.go
update_uma_go() {
    local version=$1
    local file="$PROJECT_ROOT/uma.go"
    
    if [[ ! -f "$file" ]]; then
        print_status $YELLOW "⚠️  File not found: $file"
        return
    fi
    
    # Update the Version variable
    sed -i.bak "s/var Version = \"[^\"]*\"/var Version = \"$version\"/" "$file"
    rm -f "$file.bak"
    print_status $GREEN "✅ Updated $file"
}

# Function to update version in uma.plg
update_uma_plg() {
    local version=$1
    local file="$PROJECT_ROOT/uma.plg"
    
    if [[ ! -f "$file" ]]; then
        print_status $YELLOW "⚠️  File not found: $file"
        return
    fi
    
    # Update the version entity
    sed -i.bak "s/<!ENTITY version   \"[^\"]*\">/<!ENTITY version   \"$version\">/" "$file"
    
    # Update installation message version
    sed -i.bak "s/echo \" Version: [^\"]*\"/echo \" Version: $version\"/" "$file"
    
    rm -f "$file.bak"
    print_status $GREEN "✅ Updated $file"
}

# Function to update package VERSION file
update_package_version() {
    local version=$1
    local file="$PROJECT_ROOT/package/uma/VERSION"
    
    if [[ ! -f "$file" ]]; then
        print_status $YELLOW "⚠️  File not found: $file"
        return
    fi
    
    echo "$version" > "$file"
    print_status $GREEN "✅ Updated $file"
}

# Function to update uma.page
update_uma_page() {
    local version=$1
    local file="$PROJECT_ROOT/package/uma/uma.page"
    
    if [[ ! -f "$file" ]]; then
        print_status $YELLOW "⚠️  File not found: $file"
        return
    fi
    
    # Update fallback version in PHP code
    sed -i.bak "s/echo '[^']*'/echo '$version'/" "$file"
    rm -f "$file.bak"
    print_status $GREEN "✅ Updated $file"
}

# Function to update create-plugin-package.sh
update_package_script() {
    local version=$1
    local file="$PROJECT_ROOT/package/create-plugin-package.sh"
    
    if [[ ! -f "$file" ]]; then
        print_status $YELLOW "⚠️  File not found: $file"
        return
    fi
    
    # Update VERSION variable
    sed -i.bak "s/VERSION=\"[^\"]*\"/VERSION=\"$version\"/" "$file"
    rm -f "$file.bak"
    print_status $GREEN "✅ Updated $file"
}

# Function to update test script
update_test_script() {
    local version=$1
    local file="$PROJECT_ROOT/package/test-plugin-installation.sh"
    
    if [[ ! -f "$file" ]]; then
        print_status $YELLOW "⚠️  File not found: $file"
        return
    fi
    
    # Update VERSION variable
    sed -i.bak "s/VERSION=\"[^\"]*\"/VERSION=\"$version\"/" "$file"
    rm -f "$file.bak"
    print_status $GREEN "✅ Updated $file"
}

# Function to verify all versions are consistent
verify_version_consistency() {
    local expected_version=$1
    local errors=0
    
    print_status $BLUE "🔍 Verifying version consistency..."
    
    # Check uma.go
    if [[ -f "$PROJECT_ROOT/uma.go" ]]; then
        local uma_version=$(grep "var Version = " "$PROJECT_ROOT/uma.go" | sed 's/.*"\([^"]*\)".*/\1/')
        if [[ "$uma_version" != "$expected_version" ]]; then
            print_status $RED "❌ uma.go version mismatch: $uma_version (expected: $expected_version)"
            ((errors++))
        fi
    fi
    
    # Check uma.plg
    if [[ -f "$PROJECT_ROOT/uma.plg" ]]; then
        local plg_version=$(grep "<!ENTITY version" "$PROJECT_ROOT/uma.plg" | sed 's/.*"\([^"]*\)".*/\1/')
        if [[ "$plg_version" != "$expected_version" ]]; then
            print_status $RED "❌ uma.plg version mismatch: $plg_version (expected: $expected_version)"
            ((errors++))
        fi
    fi
    
    # Check package VERSION file
    if [[ -f "$PROJECT_ROOT/package/uma/VERSION" ]]; then
        local pkg_version=$(cat "$PROJECT_ROOT/package/uma/VERSION" | tr -d '\n\r' | xargs)
        if [[ "$pkg_version" != "$expected_version" ]]; then
            print_status $RED "❌ package/uma/VERSION mismatch: $pkg_version (expected: $expected_version)"
            ((errors++))
        fi
    fi
    
    if [[ $errors -eq 0 ]]; then
        print_status $GREEN "✅ All versions are consistent: $expected_version"
        return 0
    else
        print_status $RED "❌ Found $errors version inconsistencies"
        return 1
    fi
}

# Main function
main() {
    local command=${1:-"sync"}
    
    case $command in
        "sync")
            print_status $BLUE "🔄 Starting automated version synchronization..."
            
            local version=$(get_current_version)
            print_status $BLUE "📋 Current version: $version"
            
            # Update all files
            update_uma_go "$version"
            update_uma_plg "$version"
            update_package_version "$version"
            update_uma_page "$version"
            update_package_script "$version"
            update_test_script "$version"
            
            # Verify consistency
            if verify_version_consistency "$version"; then
                print_status $GREEN "🎉 Version synchronization completed successfully!"
            else
                print_status $RED "❌ Version synchronization failed"
                exit 1
            fi
            ;;
            
        "set")
            if [[ -z "$2" ]]; then
                print_status $RED "❌ Usage: $0 set <version>"
                print_status $YELLOW "Example: $0 set 2025.06.24"
                exit 1
            fi
            
            local new_version=$2
            validate_version_format "$new_version"
            
            print_status $BLUE "🔄 Setting new version: $new_version"
            echo "$new_version" > "$VERSION_FILE"
            
            # Sync all files with new version
            $0 sync
            ;;
            
        "verify")
            local version=$(get_current_version)
            verify_version_consistency "$version"
            ;;
            
        "current")
            local version=$(get_current_version)
            echo "$version"
            ;;
            
        *)
            print_status $BLUE "UMA Automated Version Management System"
            echo ""
            echo "Usage: $0 <command> [arguments]"
            echo ""
            echo "Commands:"
            echo "  sync              Synchronize version across all files"
            echo "  set <version>     Set new version and synchronize"
            echo "  verify            Verify version consistency"
            echo "  current           Display current version"
            echo ""
            echo "Examples:"
            echo "  $0 sync                    # Sync current version"
            echo "  $0 set 2025.06.24         # Set new version"
            echo "  $0 verify                 # Check consistency"
            echo "  $0 current                # Show current version"
            ;;
    esac
}

# Run main function with all arguments
main "$@"
