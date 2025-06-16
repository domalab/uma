# Prometheus Metrics Guide

UMA provides comprehensive Prometheus metrics for monitoring API performance, system health, and operational metrics. This guide covers all available metrics and how to use them for monitoring and alerting.

## Metrics Endpoint

**URL**: `http://your-unraid-ip:34600/metrics`

The metrics endpoint exposes Prometheus-formatted metrics that can be scraped by Prometheus or compatible monitoring systems.

## Available Metrics

### API Request Metrics

#### `uma_api_requests_total`
**Type**: Counter  
**Description**: Total number of API requests  
**Labels**:
- `method` - HTTP method (GET, POST, etc.)
- `endpoint` - API endpoint path
- `status_code` - HTTP response status code

```prometheus
uma_api_requests_total{endpoint="/api/v1/health",method="GET",status_code="200"} 42
uma_api_requests_total{endpoint="/api/v1/docker/containers",method="GET",status_code="200"} 15
uma_api_requests_total{endpoint="/api/v1/docker/containers/bulk/start",method="POST",status_code="200"} 3
```

#### `uma_api_request_duration_seconds`
**Type**: Histogram  
**Description**: Duration of API requests in seconds  
**Labels**:
- `method` - HTTP method
- `endpoint` - API endpoint path

```prometheus
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.005"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.01"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.025"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.05"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.1"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.25"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.5"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="1"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="2.5"} 42
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="5"} 42
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="10"} 42
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="+Inf"} 42
uma_api_request_duration_seconds_sum{endpoint="/api/v1/health",method="GET"} 84.5
uma_api_request_duration_seconds_count{endpoint="/api/v1/health",method="GET"} 42
```

### Health Check Metrics

#### `uma_health_check_duration_seconds`
**Type**: Histogram  
**Description**: Duration of health checks in seconds

```prometheus
uma_health_check_duration_seconds_bucket{le="0.1"} 0
uma_health_check_duration_seconds_bucket{le="0.5"} 0
uma_health_check_duration_seconds_bucket{le="1"} 0
uma_health_check_duration_seconds_bucket{le="2"} 15
uma_health_check_duration_seconds_bucket{le="5"} 15
uma_health_check_duration_seconds_bucket{le="10"} 15
uma_health_check_duration_seconds_bucket{le="+Inf"} 15
uma_health_check_duration_seconds_sum 30.5
uma_health_check_duration_seconds_count 15
```

#### `uma_health_check_status`
**Type**: Gauge  
**Description**: Health check status (1 = healthy, 0 = unhealthy)  
**Labels**:
- `dependency` - Service dependency name

```prometheus
uma_health_check_status{dependency="docker"} 1
uma_health_check_status{dependency="libvirt"} 1
uma_health_check_status{dependency="storage"} 1
uma_health_check_status{dependency="notifications"} 1
```

### Bulk Operation Metrics

#### `uma_bulk_operation_duration_seconds`
**Type**: Histogram  
**Description**: Duration of bulk operations in seconds  
**Labels**:
- `operation` - Operation type (start, stop, restart)

```prometheus
uma_bulk_operation_duration_seconds_bucket{operation="start",le="0.1"} 0
uma_bulk_operation_duration_seconds_bucket{operation="start",le="0.5"} 2
uma_bulk_operation_duration_seconds_bucket{operation="start",le="1"} 3
uma_bulk_operation_duration_seconds_bucket{operation="start",le="2.5"} 3
uma_bulk_operation_duration_seconds_bucket{operation="start",le="5"} 3
uma_bulk_operation_duration_seconds_bucket{operation="start",le="10"} 3
uma_bulk_operation_duration_seconds_bucket{operation="start",le="+Inf"} 3
uma_bulk_operation_duration_seconds_sum{operation="start"} 1.85
uma_bulk_operation_duration_seconds_count{operation="start"} 3
```

#### `uma_bulk_operation_containers`
**Type**: Histogram  
**Description**: Number of containers processed in bulk operations  
**Labels**:
- `operation` - Operation type

```prometheus
uma_bulk_operation_containers_bucket{operation="start",le="1"} 1
uma_bulk_operation_containers_bucket{operation="start",le="5"} 2
uma_bulk_operation_containers_bucket{operation="start",le="10"} 3
uma_bulk_operation_containers_bucket{operation="start",le="25"} 3
uma_bulk_operation_containers_bucket{operation="start",le="50"} 3
uma_bulk_operation_containers_bucket{operation="start",le="+Inf"} 3
uma_bulk_operation_containers_sum{operation="start"} 12
uma_bulk_operation_containers_count{operation="start"} 3
```

#### `uma_bulk_operation_success_rate`
**Type**: Gauge  
**Description**: Success rate of bulk operations (percentage)  
**Labels**:
- `operation` - Operation type

```prometheus
uma_bulk_operation_success_rate{operation="start"} 100
uma_bulk_operation_success_rate{operation="stop"} 95.5
uma_bulk_operation_success_rate{operation="restart"} 98.2
```

### WebSocket Metrics

#### `uma_websocket_connections`
**Type**: Gauge  
**Description**: Number of active WebSocket connections  
**Labels**:
- `endpoint` - WebSocket endpoint

```prometheus
uma_websocket_connections{endpoint="/api/v1/ws/system/stats"} 3
uma_websocket_connections{endpoint="/api/v1/ws/docker/events"} 1
uma_websocket_connections{endpoint="/api/v1/ws/storage/status"} 2
```

#### `uma_websocket_messages_total`
**Type**: Counter  
**Description**: Total number of WebSocket messages sent  
**Labels**:
- `endpoint` - WebSocket endpoint
- `message_type` - Type of message

```prometheus
uma_websocket_messages_total{endpoint="/api/v1/ws/system/stats",message_type="stats"} 1250
uma_websocket_messages_total{endpoint="/api/v1/ws/docker/events",message_type="event"} 45
uma_websocket_messages_total{endpoint="/api/v1/ws/storage/status",message_type="status"} 180
```

### System Metrics

#### `uma_config_load_total`
**Type**: Counter  
**Description**: Total number of configuration loads  
**Labels**:
- `config_type` - Type of configuration (ini, json, etc.)
- `status` - Load status (success, error)

```prometheus
uma_config_load_total{config_type="ini",status="success"} 5
uma_config_load_total{config_type="json",status="success"} 3
uma_config_load_total{config_type="ini",status="error"} 0
```

#### `uma_validation_errors_total`
**Type**: Counter  
**Description**: Total number of validation errors  
**Labels**:
- `component` - Component that failed validation
- `operation` - Operation being validated

```prometheus
uma_validation_errors_total{component="docker",operation="bulk_start"} 2
uma_validation_errors_total{component="api",operation="pagination"} 1
```

## Prometheus Configuration

### Basic Scrape Configuration
Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'uma'
    static_configs:
      - targets: ['your-unraid-ip:34600']
    metrics_path: '/metrics'
    scrape_interval: 30s
    scrape_timeout: 10s
```

### Advanced Configuration with Labels
```yaml
scrape_configs:
  - job_name: 'uma'
    static_configs:
      - targets: ['your-unraid-ip:34600']
    metrics_path: '/metrics'
    scrape_interval: 30s
    scrape_timeout: 10s
    relabel_configs:
      - target_label: 'instance'
        replacement: 'unraid-server'
      - target_label: 'service'
        replacement: 'uma'
```

## Useful Queries

### API Performance
```promql
# Average response time by endpoint
rate(uma_api_request_duration_seconds_sum[5m]) / rate(uma_api_request_duration_seconds_count[5m])

# Request rate by endpoint
rate(uma_api_requests_total[5m])

# Error rate percentage
(rate(uma_api_requests_total{status_code!~"2.."}[5m]) / rate(uma_api_requests_total[5m])) * 100
```

### Health Monitoring
```promql
# Unhealthy dependencies
uma_health_check_status == 0

# Health check duration trend
rate(uma_health_check_duration_seconds_sum[5m]) / rate(uma_health_check_duration_seconds_count[5m])
```

### Bulk Operations
```promql
# Bulk operation success rate
uma_bulk_operation_success_rate

# Average containers per bulk operation
rate(uma_bulk_operation_containers_sum[5m]) / rate(uma_bulk_operation_containers_count[5m])

# Bulk operation duration percentiles
histogram_quantile(0.95, rate(uma_bulk_operation_duration_seconds_bucket[5m]))
```

### WebSocket Activity
```promql
# Active WebSocket connections
sum(uma_websocket_connections) by (endpoint)

# WebSocket message rate
rate(uma_websocket_messages_total[5m])
```

## Alerting Rules

### Critical Alerts
```yaml
groups:
  - name: uma.critical
    rules:
      - alert: UMAServiceDown
        expr: up{job="uma"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "UMA service is down"
          description: "UMA service has been down for more than 1 minute"

      - alert: UMAHealthCheckFailed
        expr: uma_health_check_status == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "UMA health check failed for {{ $labels.dependency }}"
          description: "Dependency {{ $labels.dependency }} has been unhealthy for more than 2 minutes"
```

### Warning Alerts
```yaml
  - name: uma.warning
    rules:
      - alert: UMAHighErrorRate
        expr: (rate(uma_api_requests_total{status_code!~"2.."}[5m]) / rate(uma_api_requests_total[5m])) * 100 > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in UMA API"
          description: "Error rate is {{ $value }}% for the last 5 minutes"

      - alert: UMASlowResponses
        expr: histogram_quantile(0.95, rate(uma_api_request_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow API responses in UMA"
          description: "95th percentile response time is {{ $value }}s"
```

## Grafana Dashboard

### Import Dashboard
1. Create a new dashboard in Grafana
2. Add panels for key metrics
3. Use the queries provided above

### Key Panels to Include
- API request rate and response times
- Health check status
- Bulk operation metrics
- WebSocket connection counts
- Error rates and alerts

### Example Panel Queries
```promql
# API Request Rate
sum(rate(uma_api_requests_total[5m])) by (endpoint)

# Response Time Heatmap
rate(uma_api_request_duration_seconds_bucket[5m])

# Health Status
uma_health_check_status

# Active WebSocket Connections
uma_websocket_connections
```

## Troubleshooting Metrics

### Metrics Not Available
1. Verify UMA is running: `curl http://your-unraid-ip:34600/api/v1/health`
2. Check metrics endpoint: `curl http://your-unraid-ip:34600/metrics`
3. Review UMA logs for errors

### Missing Specific Metrics
- Some metrics only appear after the corresponding operations occur
- Make API calls to generate metrics
- Check if the feature is enabled

### Prometheus Scraping Issues
1. Verify Prometheus can reach UMA
2. Check Prometheus logs for scrape errors
3. Validate Prometheus configuration syntax

## Next Steps

- **[Logging Guide](logging.md)** - Structured logging documentation
- **[Testing Guide](testing.md)** - Running and writing tests
- **[Monitoring Setup](../deployment/monitoring.md)** - Complete monitoring stack setup
