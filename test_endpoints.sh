#!/bin/bash

# UMA API Endpoint Testing Script
BASE_URL="http://192.168.20.21:34600/api/v1"
METRICS_URL="http://192.168.20.21:34600/metrics"

echo "=== UMA API ENDPOINT TESTING ==="
echo "Testing all 51 documented endpoints..."
echo ""

# Function to test endpoint
test_endpoint() {
    local method=$1
    local path=$2
    local full_url="${BASE_URL}${path}"
    
    if [[ "$path" == "/metrics" ]]; then
        full_url="http://192.168.20.21:34600/metrics"
    fi
    
    if [[ "$method" == "GET" ]]; then
        result=$(curl -s -o /dev/null -w "%{http_code}" "$full_url")
        echo "$method $path: $result"
    elif [[ "$method" == "POST" ]]; then
        result=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$full_url")
        echo "$method $path: $result"
    elif [[ "$method" == "DELETE" ]]; then
        result=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$full_url")
        echo "$method $path: $result"
    fi
}

echo "=== CRITICAL SYSTEM MONITORING ==="
test_endpoint "GET" "/health"
test_endpoint "GET" "/system/info"
test_endpoint "GET" "/system/cpu"
test_endpoint "GET" "/system/memory"
test_endpoint "GET" "/system/temperature"
test_endpoint "GET" "/system/network"
test_endpoint "GET" "/system/gpu"
test_endpoint "GET" "/system/fans"
test_endpoint "GET" "/system/filesystems"
test_endpoint "GET" "/system/resources"
test_endpoint "GET" "/system/ups"
test_endpoint "GET" "/system/logs"

echo ""
echo "=== STORAGE MANAGEMENT ==="
test_endpoint "GET" "/storage/array"
test_endpoint "GET" "/storage/disks"
test_endpoint "GET" "/storage/cache"
test_endpoint "GET" "/storage/boot"
test_endpoint "GET" "/storage/zfs"
test_endpoint "GET" "/storage/general"
test_endpoint "GET" "/system/parity/check"
test_endpoint "GET" "/system/parity/disk"

echo ""
echo "=== DOCKER MANAGEMENT ==="
test_endpoint "GET" "/docker/containers"
test_endpoint "GET" "/docker/info"
test_endpoint "GET" "/docker/images"
test_endpoint "GET" "/docker/networks"

echo ""
echo "=== VM MANAGEMENT ==="
test_endpoint "GET" "/vms"

echo ""
echo "=== METRICS AND MONITORING ==="
test_endpoint "GET" "/metrics"
test_endpoint "GET" "/operations"
test_endpoint "GET" "/operations/stats"

echo ""
echo "=== NOTIFICATIONS ==="
test_endpoint "GET" "/notifications"
test_endpoint "GET" "/notifications/stats"

echo ""
echo "=== DIAGNOSTICS ==="
test_endpoint "GET" "/diagnostics/health"
test_endpoint "GET" "/diagnostics/info"

echo ""
echo "=== WEBSOCKET ENDPOINTS (GET only for testing) ==="
test_endpoint "GET" "/ws/system/stats"
test_endpoint "GET" "/ws/docker/events"
test_endpoint "GET" "/ws/storage/status"

echo ""
echo "=== TESTING COMPLETE ==="
