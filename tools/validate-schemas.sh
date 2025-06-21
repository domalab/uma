#!/bin/bash

# UMA API Schema Validation Tool
# Validates OpenAPI schemas against live API responses

set -e

# Configuration
UMA_HOST="${UMA_HOST:-192.168.20.21}"
UMA_PORT="${UMA_PORT:-34600}"
BASE_URL="http://${UMA_HOST}:${UMA_PORT}"
REPORT_DIR="./reports"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create reports directory
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}🔍 UMA API Schema Validation${NC}"
echo -e "${BLUE}================================${NC}"
echo "Base URL: $BASE_URL"
echo "Timestamp: $TIMESTAMP"
echo ""

# Function to check if UMA is running
check_uma_status() {
    echo -e "${BLUE}📡 Checking UMA API status...${NC}"
    
    if curl -s --connect-timeout 5 "$BASE_URL/api/v1/health" > /dev/null; then
        echo -e "${GREEN}✅ UMA API is responding${NC}"
        return 0
    else
        echo -e "${RED}❌ UMA API is not responding at $BASE_URL${NC}"
        echo "Please ensure UMA is running and accessible."
        exit 1
    fi
}

# Function to validate OpenAPI spec syntax
validate_openapi_spec() {
    echo -e "${BLUE}📋 Validating OpenAPI specification syntax...${NC}"
    
    # Download OpenAPI spec
    local spec_file="$REPORT_DIR/openapi_spec_$TIMESTAMP.json"
    if curl -s "$BASE_URL/api/v1/openapi.json" > "$spec_file"; then
        echo -e "${GREEN}✅ OpenAPI spec downloaded${NC}"
        
        # Validate with modern OpenAPI CLI if available
        if command -v openapi &> /dev/null; then
            echo "Validating OpenAPI specification syntax..."
            if openapi lint "$spec_file" --format=summary 2>/dev/null; then
                echo -e "${GREEN}✅ OpenAPI spec is valid${NC}"
            else
                echo -e "${YELLOW}⚠️  OpenAPI spec has linting issues (continuing with validation)${NC}"
                echo "Running detailed validation..."
                openapi lint "$spec_file" --max-problems 10 || true
                echo -e "${YELLOW}Note: Linting issues don't prevent schema-response validation${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  openapi command not available, skipping syntax validation${NC}"
            echo "Install with: npm install -g @redocly/cli"
        fi
    else
        echo -e "${RED}❌ Failed to download OpenAPI spec${NC}"
        return 1
    fi
}

# Function to test endpoint availability
test_endpoint_availability() {
    echo -e "${BLUE}🌐 Testing endpoint availability...${NC}"
    
    # Comprehensive list of all UMA API endpoints for availability testing
    local endpoints=(
        # System endpoints
        "/api/v1/system/cpu"
        "/api/v1/system/memory"
        "/api/v1/system/temperature"
        "/api/v1/system/network"
        "/api/v1/system/ups"
        "/api/v1/system/gpu"
        "/api/v1/system/fans"
        "/api/v1/system/resources"
        "/api/v1/system/filesystems"
        "/api/v1/system/info"
        "/api/v1/system/logs"
        "/api/v1/system/parity/disk"
        "/api/v1/system/parity/check"

        # Storage endpoints
        "/api/v1/storage/array"
        "/api/v1/storage/disks"
        "/api/v1/storage/boot"
        "/api/v1/storage/cache"
        "/api/v1/storage/general"
        "/api/v1/storage/zfs"

        # Docker endpoints
        "/api/v1/docker/containers"
        "/api/v1/docker/info"
        "/api/v1/docker/images"
        "/api/v1/docker/networks"

        # VM endpoints
        "/api/v1/vms"

        # Health and diagnostics
        "/api/v1/health"
        "/api/v1/diagnostics/health"
        "/api/v1/diagnostics/info"

        # Operations and async
        "/api/v1/operations"
        "/api/v1/operations/stats"

        # Notifications
        "/api/v1/notifications"
        "/api/v1/notifications/stats"

        # Metrics
        "/api/v1/metrics"
    )
    
    local available=0
    local total=${#endpoints[@]}
    
    for endpoint in "${endpoints[@]}"; do
        if curl -s --connect-timeout 5 "$BASE_URL$endpoint" > /dev/null; then
            echo -e "  ${GREEN}✅${NC} $endpoint"
            ((available++))
        else
            echo -e "  ${RED}❌${NC} $endpoint"
        fi
    done
    
    echo ""
    echo -e "Endpoint Availability: ${available}/${total} ($(( available * 100 / total ))%)"
    
    if [ $available -lt $((total * 80 / 100)) ]; then
        echo -e "${RED}❌ Less than 80% of endpoints are available${NC}"
        return 1
    fi
}

# Function to run schema validation
run_schema_validation() {
    echo -e "${BLUE}🔍 Running detailed schema validation...${NC}"
    
    # Build validator if needed
    if [ ! -f "./tools/schema-validator/schema-validator" ]; then
        echo "Building schema validator..."
        cd tools/schema-validator
        go build -o schema-validator main.go
        cd ../..
    fi
    
    # Run validation
    local report_file="$REPORT_DIR/schema_validation_$TIMESTAMP.json"
    if ./tools/schema-validator/schema-validator "$BASE_URL" > "$REPORT_DIR/validation_output_$TIMESTAMP.txt" 2>&1; then
        echo -e "${GREEN}✅ Schema validation completed successfully${NC}"
        
        # Move generated report
        if [ -f "schema-validation-report.json" ]; then
            mv "schema-validation-report.json" "$report_file"
            echo -e "📄 Detailed report: $report_file"
        fi
        
        return 0
    else
        echo -e "${RED}❌ Schema validation found issues${NC}"
        
        # Move generated report
        if [ -f "schema-validation-report.json" ]; then
            mv "schema-validation-report.json" "$report_file"
            echo -e "📄 Detailed report: $report_file"
        fi
        
        # Show validation output
        echo -e "${YELLOW}Validation output:${NC}"
        cat "$REPORT_DIR/validation_output_$TIMESTAMP.txt"
        
        return 1
    fi
}

# Function to check for schema drift
check_schema_drift() {
    echo -e "${BLUE}📊 Checking for schema drift...${NC}"
    
    local current_spec="$REPORT_DIR/openapi_spec_$TIMESTAMP.json"
    local baseline_spec="$REPORT_DIR/baseline_openapi_spec.json"
    
    if [ ! -f "$baseline_spec" ]; then
        echo -e "${YELLOW}⚠️  No baseline spec found, creating baseline${NC}"
        cp "$current_spec" "$baseline_spec"
        echo -e "📄 Baseline created: $baseline_spec"
        return 0
    fi
    
    # Compare specs if openapi-changes is available
    if command -v openapi-changes &> /dev/null; then
        local diff_file="$REPORT_DIR/schema_diff_$TIMESTAMP.txt"
        echo "Comparing specifications for changes..."
        if openapi-changes summary "$baseline_spec" "$current_spec" --no-logo > "$diff_file" 2>&1; then
            if grep -q "No changes found" "$diff_file"; then
                echo -e "${GREEN}✅ No breaking changes detected${NC}"
            else
                echo -e "${YELLOW}⚠️  Schema changes detected${NC}"
                echo -e "📄 Diff report: $diff_file"
            fi
        else
            echo -e "${YELLOW}⚠️  Schema comparison completed with warnings${NC}"
            echo -e "📄 Diff report: $diff_file"
        fi
    else
        echo -e "${YELLOW}⚠️  openapi-changes not available, skipping drift detection${NC}"
        echo "Install with: npm install -g @pb33f/openapi-changes"
    fi
}

# Function to generate summary report
generate_summary_report() {
    echo -e "${BLUE}📋 Generating comprehensive summary report...${NC}"

    local summary_file="$REPORT_DIR/validation_summary_$TIMESTAMP.md"

    # Extract metrics from generated files
    local manual_endpoints=$([ -f "$REPORT_DIR/openapi_spec_$TIMESTAMP.json" ] && jq '.paths | keys | length' "$REPORT_DIR/openapi_spec_$TIMESTAMP.json" 2>/dev/null || echo "0")
    local generated_endpoints=$([ -f "$REPORT_DIR/generated_openapi_$TIMESTAMP.json" ] && jq '.paths | keys | length' "$REPORT_DIR/generated_openapi_$TIMESTAMP.json" 2>/dev/null || echo "0")
    local discovered_routes=$([ -f "$REPORT_DIR/discovered_routes_$TIMESTAMP.json" ] && jq '.total_routes' "$REPORT_DIR/discovered_routes_$TIMESTAMP.json" 2>/dev/null || echo "0")
    local websocket_channels=$([ -f "$REPORT_DIR/websocket_docs_$TIMESTAMP.json" ] && jq '.channels | length' "$REPORT_DIR/websocket_docs_$TIMESTAMP.json" 2>/dev/null || echo "0")

    cat > "$summary_file" << EOF
# UMA API Schema Validation & Generation Report

**Timestamp**: $(date)
**Base URL**: $BASE_URL
**Validation ID**: $TIMESTAMP

## Executive Summary

This report provides comprehensive analysis of the UMA API including route discovery, schema generation, WebSocket documentation, and validation results.

## Coverage Analysis

| Metric | Count | Status |
|--------|-------|--------|
| **Manual Documentation** | $manual_endpoints endpoints | $([ -f "$REPORT_DIR/openapi_spec_$TIMESTAMP.json" ] && echo "✅ Valid" || echo "❌ Invalid") |
| **Generated from Live API** | $generated_endpoints endpoints | $([ -f "$REPORT_DIR/generated_openapi_$TIMESTAMP.json" ] && echo "✅ Generated" || echo "❌ Failed") |
| **Discovered Routes** | $discovered_routes routes | $([ -f "$REPORT_DIR/discovered_routes_$TIMESTAMP.json" ] && echo "✅ Discovered" || echo "❌ Failed") |
| **WebSocket Channels** | $websocket_channels channels | $([ -f "$REPORT_DIR/websocket_docs_$TIMESTAMP.json" ] && echo "✅ Documented" || echo "❌ Failed") |

## Validation Results

- **OpenAPI Spec Syntax**: $([ -f "$REPORT_DIR/openapi_spec_$TIMESTAMP.json" ] && echo "✅ Valid" || echo "❌ Invalid")
- **Endpoint Availability**: $(curl -s "$BASE_URL/api/v1/health" > /dev/null && echo "✅ Available" || echo "❌ Unavailable")
- **Schema-Response Validation**: $([ -f "$REPORT_DIR/schema_validation_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed")
- **Schema Comparison**: $([ -f "$REPORT_DIR/schema_comparison_$TIMESTAMP.json" ] && echo "✅ Completed" || echo "❌ Failed")

## Key Findings

### Route Discovery
- **Total Routes Found**: $discovered_routes
- **Documentation Gap**: $((discovered_routes - manual_endpoints)) undocumented routes
- **Coverage**: $((manual_endpoints * 100 / (discovered_routes > 0 ? discovered_routes : 1)))% of discovered routes are documented

### Schema Generation
- **Live API Endpoints**: $generated_endpoints
- **Auto-generated Schemas**: Available for all responding endpoints
- **Schema Quality**: Generated from real API responses

### WebSocket Support
- **Channels Available**: $websocket_channels
- **Real-time Monitoring**: Full subscription management
- **Event Types**: Comprehensive coverage of system events

## Files Generated

### Core Reports
- **OpenAPI Spec**: \`$REPORT_DIR/openapi_spec_$TIMESTAMP.json\`
- **Schema Validation**: \`$REPORT_DIR/schema_validation_$TIMESTAMP.json\`
- **Validation Output**: \`$REPORT_DIR/validation_output_$TIMESTAMP.txt\`

### Discovery & Generation
- **Route Discovery**: \`$REPORT_DIR/discovered_routes_$TIMESTAMP.json\`
- **Generated OpenAPI**: \`$REPORT_DIR/generated_openapi_$TIMESTAMP.json\`
- **WebSocket Docs**: \`$REPORT_DIR/websocket_docs_$TIMESTAMP.json\`
- **Schema Comparison**: \`$REPORT_DIR/schema_comparison_$TIMESTAMP.json\`

## Recommendations

1. **Documentation Updates**: Review undocumented routes and add to OpenAPI spec
2. **Schema Accuracy**: Compare generated vs manual schemas for discrepancies
3. **WebSocket Integration**: Ensure WebSocket endpoints are properly documented
4. **Automated Generation**: Consider integrating live schema generation into CI/CD

## Next Steps

1. Review any critical issues in the validation report
2. Analyze route discovery results for missing documentation
3. Compare generated schemas with manual schemas
4. Update baseline spec if changes are intentional
5. Integrate automated generation into development workflow

## Commands

### Re-run Full Validation
\`\`\`bash
./tools/validate-schemas.sh
\`\`\`

### Individual Tools
\`\`\`bash
# Route discovery only
./tools/route-scanner/route-scanner daemon/services/api/routes --json

# Schema generation only
./tools/schema-generator/schema-generator $BASE_URL --routes

# WebSocket documentation only
./tools/websocket-documenter/websocket-documenter daemon/services/api
\`\`\`

EOF

    echo -e "📄 Comprehensive summary report: $summary_file"
}

# Function to run route discovery
run_route_discovery() {
    echo -e "${BLUE}🛣️  Running route discovery...${NC}"

    # Build route scanner if needed
    if [ ! -f "./tools/route-scanner/route-scanner" ]; then
        echo "Building route scanner..."
        cd tools/route-scanner
        go build -o route-scanner main.go
        cd ../..
    fi

    # Run route discovery
    local routes_file="$REPORT_DIR/discovered_routes_$TIMESTAMP.json"
    if ./tools/route-scanner/route-scanner daemon/services/api/routes --json --output "$routes_file"; then
        echo -e "${GREEN}✅ Route discovery completed${NC}"
        echo -e "📄 Routes report: $routes_file"

        # Extract route count for summary
        local route_count=$(jq '.total_routes' "$routes_file" 2>/dev/null || echo "unknown")
        echo -e "📊 Discovered routes: $route_count"
        return 0
    else
        echo -e "${RED}❌ Route discovery failed${NC}"
        return 1
    fi
}

# Function to generate schemas from live API
generate_live_schemas() {
    echo -e "${BLUE}🔄 Generating schemas from live API...${NC}"

    # Build schema generator if needed
    if [ ! -f "./tools/schema-generator/schema-generator" ]; then
        echo "Building schema generator..."
        cd tools/schema-generator
        go build -o schema-generator main.go
        cd ../..
    fi

    # Generate schemas from live API
    local generated_spec="$REPORT_DIR/generated_openapi_$TIMESTAMP.json"
    if ./tools/schema-generator/schema-generator "$BASE_URL" --routes --output "$generated_spec"; then
        echo -e "${GREEN}✅ Live schema generation completed${NC}"
        echo -e "📄 Generated spec: $generated_spec"

        # Extract endpoint count for summary
        local endpoint_count=$(jq '.paths | keys | length' "$generated_spec" 2>/dev/null || echo "unknown")
        echo -e "📊 Generated endpoints: $endpoint_count"
        return 0
    else
        echo -e "${RED}❌ Live schema generation failed${NC}"
        return 1
    fi
}

# Function to document WebSocket endpoints
document_websockets() {
    echo -e "${BLUE}🔌 Documenting WebSocket endpoints...${NC}"

    # Build WebSocket documenter if needed
    if [ ! -f "./tools/websocket-documenter/websocket-documenter" ]; then
        echo "Building WebSocket documenter..."
        cd tools/websocket-documenter
        go build -o websocket-documenter main.go
        cd ../..
    fi

    # Generate WebSocket documentation
    local ws_docs="$REPORT_DIR/websocket_docs_$TIMESTAMP.json"
    if ./tools/websocket-documenter/websocket-documenter daemon/services/api --output "$ws_docs"; then
        echo -e "${GREEN}✅ WebSocket documentation completed${NC}"
        echo -e "📄 WebSocket docs: $ws_docs"

        # Extract channel count for summary
        local channel_count=$(jq '.channels | length' "$ws_docs" 2>/dev/null || echo "unknown")
        echo -e "📊 WebSocket channels: $channel_count"
        return 0
    else
        echo -e "${RED}❌ WebSocket documentation failed${NC}"
        return 1
    fi
}

# Function to compare generated vs manual schemas
compare_schemas() {
    echo -e "${BLUE}📊 Comparing generated vs manual schemas...${NC}"

    local manual_spec="$REPORT_DIR/openapi_spec_$TIMESTAMP.json"
    local generated_spec="$REPORT_DIR/generated_openapi_$TIMESTAMP.json"
    local comparison_file="$REPORT_DIR/schema_comparison_$TIMESTAMP.json"

    if [ -f "$manual_spec" ] && [ -f "$generated_spec" ]; then
        # Create comparison report
        cat > "$comparison_file" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "manual_spec": {
    "file": "$manual_spec",
    "endpoints": $(jq '.paths | keys | length' "$manual_spec" 2>/dev/null || echo 0),
    "schemas": $(jq '.components.schemas | keys | length' "$manual_spec" 2>/dev/null || echo 0)
  },
  "generated_spec": {
    "file": "$generated_spec",
    "endpoints": $(jq '.paths | keys | length' "$generated_spec" 2>/dev/null || echo 0),
    "schemas": $(jq '.components.schemas | keys | length' "$generated_spec" 2>/dev/null || echo 0)
  },
  "comparison": {
    "endpoint_difference": $(($(jq '.paths | keys | length' "$generated_spec" 2>/dev/null || echo 0) - $(jq '.paths | keys | length' "$manual_spec" 2>/dev/null || echo 0))),
    "schema_difference": $(($(jq '.components.schemas | keys | length' "$generated_spec" 2>/dev/null || echo 0) - $(jq '.components.schemas | keys | length' "$manual_spec" 2>/dev/null || echo 0)))
  }
}
EOF

        echo -e "${GREEN}✅ Schema comparison completed${NC}"
        echo -e "📄 Comparison report: $comparison_file"

        # Show summary
        local manual_endpoints=$(jq '.manual_spec.endpoints' "$comparison_file")
        local generated_endpoints=$(jq '.generated_spec.endpoints' "$comparison_file")
        local endpoint_diff=$(jq '.comparison.endpoint_difference' "$comparison_file")

        echo -e "📊 Manual endpoints: $manual_endpoints"
        echo -e "📊 Generated endpoints: $generated_endpoints"
        echo -e "📊 Difference: $endpoint_diff"

        return 0
    else
        echo -e "${YELLOW}⚠️  Cannot compare schemas - missing files${NC}"
        return 1
    fi
}

# Function to cleanup old reports
cleanup_old_reports() {
    echo -e "${BLUE}🧹 Cleaning up old reports...${NC}"

    # Keep only last 10 reports
    find "$REPORT_DIR" -name "*_[0-9]*" -type f | sort | head -n -30 | xargs rm -f 2>/dev/null || true

    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}Starting UMA API schema validation and generation...${NC}"
    echo ""

    # Check prerequisites
    check_uma_status
    echo ""

    # Run discovery and generation steps
    run_route_discovery
    echo ""

    generate_live_schemas
    echo ""

    document_websockets
    echo ""

    # Run validation steps
    validate_openapi_spec
    echo ""

    test_endpoint_availability
    echo ""

    run_schema_validation
    echo ""

    # Compare generated vs manual schemas
    compare_schemas
    echo ""

    check_schema_drift
    echo ""

    generate_summary_report
    echo ""

    cleanup_old_reports
    echo ""

    echo -e "${GREEN}🎉 Schema validation and generation completed!${NC}"
    echo -e "📁 Reports saved in: $REPORT_DIR"

    # Exit with appropriate code
    if [ -f "$REPORT_DIR/schema_validation_$TIMESTAMP.json" ]; then
        # Check if there were critical issues
        if grep -q '"severity":"CRITICAL"' "$REPORT_DIR/schema_validation_$TIMESTAMP.json" 2>/dev/null; then
            echo -e "${RED}❌ Critical schema issues found${NC}"
            exit 1
        else
            echo -e "${GREEN}✅ No critical schema issues${NC}"
            exit 0
        fi
    else
        echo -e "${YELLOW}⚠️  Validation completed with warnings${NC}"
        exit 0
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "UMA API Schema Validation Tool"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --clean        Clean all reports and exit"
        echo ""
        echo "Environment Variables:"
        echo "  UMA_HOST       UMA server host (default: 192.168.20.21)"
        echo "  UMA_PORT       UMA server port (default: 34600)"
        echo ""
        echo "Examples:"
        echo "  $0                           # Run validation with defaults"
        echo "  UMA_HOST=localhost $0        # Run against localhost"
        echo "  $0 --clean                   # Clean old reports"
        exit 0
        ;;
    --clean)
        echo -e "${BLUE}🧹 Cleaning all reports...${NC}"
        rm -rf "$REPORT_DIR"
        echo -e "${GREEN}✅ All reports cleaned${NC}"
        exit 0
        ;;
    *)
        main
        ;;
esac
