#!/bin/bash
#
# UMA Plugin Installation Test Script
# Tests the complete plugin installation and functionality
#

set -e

# Configuration
PLUGIN_NAME="uma"
VERSION="2025.06.24"
TEST_HOST="${UMA_TEST_HOST:-192.168.20.21}"
TEST_PASSWORD="${UMA_TEST_PASSWORD:-tasvyh-4Gehju-ridxic}"

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

# Function to execute remote command
remote_exec() {
    sshpass -p "$TEST_PASSWORD" ssh -o StrictHostKeyChecking=no root@"$TEST_HOST" "$1"
}

# Function to copy file to remote
remote_copy() {
    sshpass -p "$TEST_PASSWORD" scp -o StrictHostKeyChecking=no "$1" root@"$TEST_HOST":"$2"
}

# Function to test prerequisites
test_prerequisites() {
    log_info "Testing prerequisites..."
    
    # Check if sshpass is available
    if ! command -v sshpass >/dev/null 2>&1; then
        log_error "sshpass is required for remote testing"
        exit 1
    fi
    
    # Test connection to Unraid server
    if ! remote_exec "echo 'Connection test successful'"; then
        log_error "Cannot connect to Unraid server at $TEST_HOST"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to test plugin installation
test_plugin_installation() {
    log_info "Testing plugin installation..."
    
    # Check if plugin files exist
    local plugin_dir="/usr/local/emhttp/plugins/$PLUGIN_NAME"
    
    if ! remote_exec "[ -d '$plugin_dir' ]"; then
        log_error "Plugin directory not found: $plugin_dir"
        return 1
    fi
    
    # Check if binary exists and is executable
    if ! remote_exec "[ -x '$plugin_dir/uma' ]"; then
        log_error "UMA binary not found or not executable"
        return 1
    fi
    
    # Check if web interface exists
    if ! remote_exec "[ -f '$plugin_dir/uma.page' ]"; then
        log_error "Web interface file not found"
        return 1
    fi
    
    # Check if scripts exist
    local scripts=("start" "stop" "restart" "status")
    for script in "${scripts[@]}"; do
        if ! remote_exec "[ -x '$plugin_dir/scripts/$script' ]"; then
            log_error "Script not found or not executable: $script"
            return 1
        fi
    done
    
    # Check if event handlers exist
    local events=("started" "stopping_svcs")
    for event in "${events[@]}"; do
        if ! remote_exec "[ -x '$plugin_dir/event/$event' ]"; then
            log_error "Event handler not found or not executable: $event"
            return 1
        fi
    done
    
    log_success "Plugin installation test passed"
    return 0
}

# Function to test plugin registration
test_plugin_registration() {
    log_info "Testing plugin registration..."
    
    # Check if plugin is registered in Unraid system
    if remote_exec "[ -L '/var/log/plugins/$PLUGIN_NAME' ]"; then
        log_success "Plugin is properly registered with Unraid"
        return 0
    else
        log_warning "Plugin symlink not found in /var/log/plugins/"
        log_info "This is expected if plugin was installed manually"
        return 0
    fi
}

# Function to test configuration
test_configuration() {
    log_info "Testing configuration..."
    
    local config_dir="/boot/config/plugins/$PLUGIN_NAME"
    
    # Check if configuration directory exists
    if ! remote_exec "[ -d '$config_dir' ]"; then
        log_error "Configuration directory not found: $config_dir"
        return 1
    fi
    
    # Check if configuration files exist
    if ! remote_exec "[ -f '$config_dir/uma.cfg' ]"; then
        log_error "Configuration file not found: uma.cfg"
        return 1
    fi
    
    if ! remote_exec "[ -f '$config_dir/uma.json' ]"; then
        log_error "JSON configuration file not found: uma.json"
        return 1
    fi
    
    # Test configuration parsing
    local service_status
    service_status=$(remote_exec "source '$config_dir/uma.cfg' && echo \$SERVICE")
    
    if [ -z "$service_status" ]; then
        log_error "Failed to parse configuration file"
        return 1
    fi
    
    log_success "Configuration test passed (SERVICE=$service_status)"
    return 0
}

# Function to test service management
test_service_management() {
    log_info "Testing service management..."
    
    local script_dir="/usr/local/emhttp/plugins/$PLUGIN_NAME/scripts"
    
    # Test stop script (in case service is running)
    log_info "Testing stop script..."
    if remote_exec "$script_dir/stop"; then
        log_success "Stop script executed successfully"
    else
        log_warning "Stop script failed (service may not have been running)"
    fi
    
    # Wait a moment
    sleep 2
    
    # Test start script
    log_info "Testing start script..."
    if remote_exec "$script_dir/start"; then
        log_success "Start script executed successfully"
    else
        log_error "Start script failed"
        return 1
    fi
    
    # Wait for service to start
    sleep 8
    
    # Test status script
    log_info "Testing status script..."
    if remote_exec "$script_dir/status"; then
        log_success "Status script executed successfully"
    else
        log_error "Status script failed"
        return 1
    fi
    
    # Test restart script
    log_info "Testing restart script..."
    if remote_exec "$script_dir/restart"; then
        log_success "Restart script executed successfully"
    else
        log_error "Restart script failed"
        return 1
    fi
    
    log_success "Service management test passed"
    return 0
}

# Function to test API functionality
test_api_functionality() {
    log_info "Testing API functionality..."
    
    # Wait for service to be ready
    sleep 5
    
    # Test health endpoint
    local health_response
    health_response=$(remote_exec "curl -s http://localhost:34600/api/v1/health")
    
    if [ -z "$health_response" ]; then
        log_error "API health endpoint not responding"
        return 1
    fi
    
    # Test if response contains expected data
    if echo "$health_response" | grep -q '"status"'; then
        log_success "API health endpoint responding correctly"
    else
        log_error "API health endpoint response invalid"
        return 1
    fi
    
    # Test a few more endpoints
    local endpoints=("system/info" "docker/containers" "storage/array" "vms")
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: /api/v1/$endpoint"
        if remote_exec "curl -s http://localhost:34600/api/v1/$endpoint >/dev/null"; then
            log_success "Endpoint $endpoint responding"
        else
            log_warning "Endpoint $endpoint not responding (may be expected)"
        fi
    done
    
    log_success "API functionality test passed"
    return 0
}

# Function to test MCP functionality
test_mcp_functionality() {
    log_info "Testing MCP functionality..."
    
    # Test MCP status endpoint
    local mcp_response
    mcp_response=$(remote_exec "curl -s http://localhost:34600/api/v1/mcp/status")
    
    if [ -z "$mcp_response" ]; then
        log_error "MCP status endpoint not responding"
        return 1
    fi
    
    if echo "$mcp_response" | grep -q '"success"'; then
        log_success "MCP status endpoint responding correctly"
    else
        log_error "MCP status endpoint response invalid"
        return 1
    fi
    
    # Test MCP configuration endpoint
    if remote_exec "curl -s http://localhost:34600/api/v1/mcp/config >/dev/null"; then
        log_success "MCP configuration endpoint responding"
    else
        log_error "MCP configuration endpoint not responding"
        return 1
    fi
    
    log_success "MCP functionality test passed"
    return 0
}

# Function to display test summary
display_test_summary() {
    echo
    echo "=== UMA Plugin Installation Test Summary ==="
    echo "Plugin: $PLUGIN_NAME"
    echo "Version: $VERSION"
    echo "Test Host: $TEST_HOST"
    echo
    log_success "All tests completed successfully!"
    echo
    echo "The UMA plugin is properly installed and functional."
    echo "You can access the web interface through Unraid's plugin settings."
    echo "API is available at: http://$TEST_HOST:34600/api/v1/"
}

# Main execution
main() {
    echo "=== UMA Plugin Installation Test ==="
    echo "Testing plugin installation on: $TEST_HOST"
    echo
    
    test_prerequisites
    test_plugin_installation
    test_plugin_registration
    test_configuration
    test_service_management
    test_api_functionality
    test_mcp_functionality
    
    display_test_summary
}

# Execute main function
main "$@"
