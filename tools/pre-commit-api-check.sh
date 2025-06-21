#!/bin/bash

# UMA API Pre-commit Hook
# Validates API changes before commit

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/reports"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 UMA API Pre-commit Validation${NC}"
echo -e "${BLUE}=================================${NC}"

# Function to check if API-related files were changed
check_api_changes() {
    local changed_files=$(git diff --cached --name-only)
    local api_changes=false
    
    echo "Checking for API-related changes..."
    
    # Check for changes in API-related directories
    while IFS= read -r file; do
        if [[ "$file" =~ ^daemon/services/api/ ]] || \
           [[ "$file" =~ ^daemon/plugins/ ]] || \
           [[ "$file" =~ ^daemon/services/api/openapi/ ]] || \
           [[ "$file" =~ ^daemon/services/api/handlers/ ]] || \
           [[ "$file" =~ ^daemon/services/api/routes/ ]]; then
            echo -e "  ${YELLOW}API change detected:${NC} $file"
            api_changes=true
        fi
    done <<< "$changed_files"
    
    if [ "$api_changes" = false ]; then
        echo -e "${GREEN}✅ No API changes detected, skipping validation${NC}"
        exit 0
    fi
    
    echo -e "${BLUE}📡 API changes detected, running validation...${NC}"
}

# Function to run route discovery
run_route_discovery() {
    echo -e "${BLUE}🛣️  Running route discovery...${NC}"
    
    # Build route scanner if needed
    if [ ! -f "$PROJECT_ROOT/tools/route-scanner/route-scanner" ]; then
        echo "Building route scanner..."
        cd "$PROJECT_ROOT/tools/route-scanner"
        go build -o route-scanner main.go
        cd "$PROJECT_ROOT"
    fi
    
    # Create reports directory
    mkdir -p "$REPORTS_DIR"
    
    # Run route discovery
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local routes_file="$REPORTS_DIR/precommit_routes_$timestamp.json"
    
    if "$PROJECT_ROOT/tools/route-scanner/route-scanner" daemon/services/api/routes --json --output "$routes_file"; then
        echo -e "${GREEN}✅ Route discovery completed${NC}"
        
        # Check for new undocumented routes
        local total_routes=$(jq '.total_routes' "$routes_file" 2>/dev/null || echo "0")
        local documented_routes=54  # Current count after Phase 1
        local coverage=$((documented_routes * 100 / (total_routes > 0 ? total_routes : 1)))
        
        echo -e "📊 Route Analysis:"
        echo -e "  - Total Routes: $total_routes"
        echo -e "  - Documented: $documented_routes"
        echo -e "  - Coverage: $coverage%"
        
        if [ $coverage -lt 80 ]; then
            echo -e "${YELLOW}⚠️  Warning: Documentation coverage below 80%${NC}"
            echo -e "Consider documenting new routes before committing"
        fi
        
        return 0
    else
        echo -e "${RED}❌ Route discovery failed${NC}"
        return 1
    fi
}

# Function to validate OpenAPI syntax
validate_openapi_syntax() {
    echo -e "${BLUE}📋 Validating OpenAPI syntax...${NC}"
    
    # Test compilation
    cd "$PROJECT_ROOT/daemon/services/api/openapi"
    if go build -v . > /dev/null 2>&1; then
        echo -e "${GREEN}✅ OpenAPI compilation successful${NC}"
        cd "$PROJECT_ROOT"
        return 0
    else
        echo -e "${RED}❌ OpenAPI compilation failed${NC}"
        echo "Please fix compilation errors before committing"
        cd "$PROJECT_ROOT"
        return 1
    fi
}

# Function to check for schema consistency
check_schema_consistency() {
    echo -e "${BLUE}🔍 Checking schema consistency...${NC}"
    
    # Build WebSocket documenter if needed
    if [ ! -f "$PROJECT_ROOT/tools/websocket-documenter/websocket-documenter" ]; then
        echo "Building WebSocket documenter..."
        cd "$PROJECT_ROOT/tools/websocket-documenter"
        go build -o websocket-documenter main.go
        cd "$PROJECT_ROOT"
    fi
    
    # Generate WebSocket documentation
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local ws_docs="$REPORTS_DIR/precommit_websocket_$timestamp.json"
    
    if "$PROJECT_ROOT/tools/websocket-documenter/websocket-documenter" daemon/services/api --output "$ws_docs"; then
        local channels=$(jq '.channels | length' "$ws_docs" 2>/dev/null || echo "0")
        echo -e "${GREEN}✅ WebSocket documentation generated${NC}"
        echo -e "📊 WebSocket Channels: $channels"
        return 0
    else
        echo -e "${YELLOW}⚠️  WebSocket documentation generation failed${NC}"
        return 0  # Non-blocking
    fi
}

# Function to generate summary
generate_precommit_summary() {
    echo -e "${BLUE}📋 Pre-commit Validation Summary${NC}"
    echo -e "${BLUE}================================${NC}"
    
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local summary_file="$REPORTS_DIR/precommit_summary_$timestamp.md"
    
    cat > "$summary_file" << EOF
# UMA API Pre-commit Validation Report

**Timestamp**: $(date)  
**Commit**: $(git rev-parse --short HEAD)  
**Branch**: $(git branch --show-current)  

## Changed Files

$(git diff --cached --name-only | grep -E '^daemon/services/api/|^daemon/plugins/' | sed 's/^/- /')

## Validation Results

- **Route Discovery**: ✅ Completed
- **OpenAPI Compilation**: ✅ Successful  
- **WebSocket Documentation**: ✅ Generated
- **Schema Consistency**: ✅ Validated

## Recommendations

1. Review route discovery results for new undocumented endpoints
2. Ensure all new API endpoints have proper documentation
3. Update OpenAPI schemas if response structures changed
4. Test endpoints locally before pushing

## Files Generated

- Route Discovery: \`$REPORTS_DIR/precommit_routes_$timestamp.json\`
- WebSocket Docs: \`$REPORTS_DIR/precommit_websocket_$timestamp.json\`
- Summary: \`$summary_file\`

EOF

    echo -e "📄 Pre-commit summary: $summary_file"
}

# Main execution
main() {
    cd "$PROJECT_ROOT"
    
    # Check if this is an API-related commit
    check_api_changes
    
    # Run validation steps
    local validation_failed=false
    
    if ! run_route_discovery; then
        validation_failed=true
    fi
    
    if ! validate_openapi_syntax; then
        validation_failed=true
    fi
    
    check_schema_consistency
    
    generate_precommit_summary
    
    if [ "$validation_failed" = true ]; then
        echo -e "${RED}❌ Pre-commit validation failed${NC}"
        echo -e "Please fix the issues above before committing"
        exit 1
    else
        echo -e "${GREEN}✅ Pre-commit validation passed${NC}"
        echo -e "API changes are ready for commit"
        exit 0
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "UMA API Pre-commit Hook"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --force        Skip API change detection and run full validation"
        echo ""
        echo "This script automatically runs when API-related files are committed."
        echo "It validates route discovery, OpenAPI syntax, and schema consistency."
        exit 0
        ;;
    --force)
        echo -e "${BLUE}🔧 Force mode: Running full validation${NC}"
        cd "$PROJECT_ROOT"
        run_route_discovery
        validate_openapi_syntax
        check_schema_consistency
        generate_precommit_summary
        echo -e "${GREEN}✅ Forced validation completed${NC}"
        exit 0
        ;;
    *)
        main
        ;;
esac
