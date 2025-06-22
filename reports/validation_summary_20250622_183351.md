# UMA API Schema Validation & Generation Report

**Timestamp**: Sun Jun 22 18:34:19 AEST 2025
**Base URL**: http://192.168.20.21:34600
**Validation ID**: 20250622_183351

## Executive Summary

This report provides comprehensive analysis of the UMA API including route discovery, schema generation, WebSocket documentation, and validation results.

## Coverage Analysis

| Metric | Count | Status |
|--------|-------|--------|
| **Manual Documentation** | 7 endpoints | ✅ Valid |
| **Generated from Live API** | 60 endpoints | ✅ Generated |
| **Discovered Routes** | 75 routes | ✅ Discovered |
| **WebSocket Channels** | 3 channels | ✅ Documented |

## Validation Results

- **OpenAPI Spec Syntax**: ✅ Valid
- **Endpoint Availability**: ✅ Available
- **Schema-Response Validation**: ✅ Completed
- **Schema Comparison**: ✅ Completed

## Key Findings

### Route Discovery
- **Total Routes Found**: 75
- **Documentation Gap**: 68 undocumented routes
- **Coverage**: 9% of discovered routes are documented

### Schema Generation
- **Live API Endpoints**: 60
- **Auto-generated Schemas**: Available for all responding endpoints
- **Schema Quality**: Generated from real API responses

### WebSocket Support
- **Channels Available**: 3
- **Real-time Monitoring**: Full subscription management
- **Event Types**: Comprehensive coverage of system events

## Files Generated

### Core Reports
- **OpenAPI Spec**: `./reports/openapi_spec_20250622_183351.json`
- **Schema Validation**: `./reports/schema_validation_20250622_183351.json`
- **Validation Output**: `./reports/validation_output_20250622_183351.txt`

### Discovery & Generation
- **Route Discovery**: `./reports/discovered_routes_20250622_183351.json`
- **Generated OpenAPI**: `./reports/generated_openapi_20250622_183351.json`
- **WebSocket Docs**: `./reports/websocket_docs_20250622_183351.json`
- **Schema Comparison**: `./reports/schema_comparison_20250622_183351.json`

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
```bash
./tools/validate-schemas.sh
```

### Individual Tools
```bash
# Route discovery only
./tools/route-scanner/route-scanner daemon/services/api/routes --json

# Schema generation only
./tools/schema-generator/schema-generator http://192.168.20.21:34600 --routes

# WebSocket documentation only
./tools/websocket-documenter/websocket-documenter daemon/services/api
```

