# Structured Logging Guide

UMA uses Zerolog for structured logging, providing production-grade logging with contextual fields, multiple output formats, and efficient performance. This guide covers the logging system and how to use it effectively.

## Logging Framework

**Library**: [Zerolog](https://github.com/rs/zerolog)  
**Format**: Structured JSON with human-readable console output  
**Levels**: Debug, Info, Warn, Error, Fatal  
**Features**: Contextual fields, request tracking, component identification

## Log Levels

### Debug
Development and troubleshooting information:
```
09:00:36 DBG Metric collected component=metrics label_endpoint=/metrics metric_type=api_request service=uma value=0.000658523
```

### Info
Normal operational messages:
```
09:00:56 INF API request completed component=api duration=1042.528298 method=GET path=/api/v1/storage/disks request_id= service=uma status_code=200
```

### Warn
Warning conditions that should be monitored:
```
09:01:15 WRN Input validation failed component=docker operation=bulk_start request_id=test-123 validation_error="duplicate container ID: plex"
```

### Error
Error conditions that need attention:
```
09:01:30 ERR Configuration load failed component=config config_type=ini error="file not found" path=/etc/uma/config.ini success=false
```

### Fatal
Critical errors that cause the application to exit:
```
09:01:45 FTL Unable to start HTTP server error="port already in use" port=34600
```

## Structured Fields

### Standard Fields
All log entries include these standard fields:
- `service=uma` - Service identifier
- `timestamp` - ISO 8601 timestamp
- `level` - Log level (DBG, INF, WRN, ERR, FTL)

### Component Fields
Logs are categorized by component:
- `component=api` - API request handling
- `component=docker` - Docker operations
- `component=health` - Health checks
- `component=websocket` - WebSocket connections
- `component=metrics` - Metrics collection
- `component=config` - Configuration loading

### Request Tracking
API requests include tracking fields:
- `request_id` - Unique request identifier
- `method` - HTTP method (GET, POST, etc.)
- `path` - API endpoint path
- `status_code` - HTTP response status
- `duration` - Request duration in milliseconds

### Operation Context
Specific operations include relevant context:
- `operation` - Operation type (bulk_start, health_check, etc.)
- `total` - Total items processed
- `succeeded` - Successful operations
- `failed` - Failed operations
- `success_rate` - Success percentage

## Log Examples

### API Request Logging
```
09:00:56 INF API request completed component=api duration=1042.528298 method=GET path=/api/v1/storage/disks request_id=storage-query-123 service=uma status_code=200
```

### Bulk Operation Logging
```
09:01:10 INF Bulk operation completed component=docker duration=175.416802 failed=0 operation=start operation_type=bulk request_id=bulk-start-456 service=uma succeeded=1 success_rate=100 total=1
```

### Health Check Logging
```
09:01:25 INF Health check completed component=health dep_docker=healthy dep_libvirt=healthy dep_notifications=healthy dep_storage=healthy duration=2394.89002 request_id=health-check-789 service=uma status=healthy
```

### WebSocket Connection Logging
```
09:01:40 INF WebSocket connection event active_connections=3 client_id=client_1671234567 component=websocket endpoint=/api/v1/ws/system/stats event=connect service=uma
```

### Validation Error Logging
```
09:01:55 WRN Input validation failed component=docker operation=bulk_start request_id=validation-test-101 service=uma validation_error="duplicate container ID: plex"
```

### Configuration Loading
```
09:02:10 INF Configuration loaded component=config config_type=ini path=/etc/uma/config.ini service=uma success=true
```

### Metrics Collection
```
09:02:25 DBG Metric collected component=metrics label_endpoint=/api/v1/health label_method=GET label_status_code=200 metric_type=api_request service=uma value=2.3951897669999997
```

## Log Configuration

### Console Output
Human-readable format for development:
```
09:00:56 INF API request completed component=api duration=1042.528 method=GET path=/api/v1/storage/disks
```

### JSON Output
Machine-readable format for production:
```json
{
  "level": "info",
  "service": "uma",
  "component": "api",
  "method": "GET",
  "path": "/api/v1/storage/disks",
  "status_code": 200,
  "duration": 1042.528298,
  "request_id": "storage-query-123",
  "time": "2025-06-15T09:00:56Z",
  "message": "API request completed"
}
```

## Using Logs for Monitoring

### Log Aggregation
Collect logs using tools like:
- **Fluentd** - Log collection and forwarding
- **Logstash** - Log processing and transformation
- **Promtail** - Grafana Loki log collection
- **Vector** - High-performance log collection

### Log Analysis
Analyze logs with:
- **Elasticsearch + Kibana** - Full-text search and visualization
- **Grafana Loki** - Log aggregation and querying
- **Splunk** - Enterprise log analysis
- **CloudWatch Logs** - AWS log management

### Example Fluentd Configuration
```xml
<source>
  @type tail
  path /var/log/uma/uma.log
  pos_file /var/log/fluentd/uma.log.pos
  tag uma.logs
  format json
  time_key time
  time_format %Y-%m-%dT%H:%M:%SZ
</source>

<match uma.logs>
  @type elasticsearch
  host elasticsearch.example.com
  port 9200
  index_name uma-logs
  type_name _doc
</match>
```

### Example Loki Configuration
```yaml
# promtail.yml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: uma
    static_configs:
      - targets:
          - localhost
        labels:
          job: uma
          __path__: /var/log/uma/uma.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            component: component
            request_id: request_id
            method: method
            path: path
      - labels:
          level:
          component:
          method:
```

## Log Queries

### Grafana Loki Queries
```logql
# All API requests
{job="uma"} |= "API request completed"

# Error logs only
{job="uma"} | json | level="error"

# Specific component logs
{job="uma"} | json | component="docker"

# Slow requests (duration > 1000ms)
{job="uma"} | json | duration > 1000

# Bulk operations
{job="uma"} | json | operation_type="bulk"

# Failed operations
{job="uma"} | json | succeeded < total
```

### Elasticsearch Queries
```json
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service": "uma"}},
        {"term": {"component": "api"}},
        {"range": {"duration": {"gte": 1000}}}
      ]
    }
  }
}
```

## Alerting on Logs

### Loki Alerting Rules
```yaml
groups:
  - name: uma-logs
    rules:
      - alert: UMAHighErrorRate
        expr: |
          (
            sum(rate({job="uma"} | json | level="error" [5m]))
            /
            sum(rate({job="uma"} [5m]))
          ) > 0.05
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in UMA logs"
          description: "Error rate is above 5% for the last 5 minutes"

      - alert: UMASlowRequests
        expr: |
          sum(rate({job="uma"} | json | duration > 5000 [5m])) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow requests detected in UMA"
          description: "More than 0.1 requests per second are taking longer than 5 seconds"
```

### ElastAlert Rules
```yaml
# elastalert_rules/uma_errors.yml
name: UMA Error Rate
type: frequency
index: uma-logs-*
num_events: 10
timeframe:
  minutes: 5

filter:
- term:
    service: "uma"
- term:
    level: "error"

alert:
- "email"

email:
- "admin@example.com"

alert_text: |
  UMA has generated {0} error logs in the last 5 minutes.
  
  Check the logs for details:
  {1}
```

## Log Rotation

### Logrotate Configuration
```bash
# /etc/logrotate.d/uma
/var/log/uma/uma.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 uma uma
    postrotate
        systemctl reload uma
    endscript
}
```

### Docker Logging
```yaml
# docker-compose.yml
version: '3.8'
services:
  uma:
    image: uma:latest
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Performance Considerations

### Log Volume
- **Debug logs**: High volume, disable in production
- **Info logs**: Moderate volume, essential for monitoring
- **Warn/Error logs**: Low volume, always enabled

### Sampling
For high-traffic environments, consider log sampling:
```go
// Sample 10% of debug logs
if rand.Float64() < 0.1 {
    logger.Debug().Msg("Debug message")
}
```

### Asynchronous Logging
Zerolog supports asynchronous logging for high-performance scenarios:
```go
logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
    With().
    Timestamp().
    Logger().
    Sample(&zerolog.BasicSampler{N: 10}) // Sample every 10th log
```

## Troubleshooting

### Common Issues

**Logs not appearing:**
- Check log level configuration
- Verify log output destination
- Ensure proper permissions

**Performance impact:**
- Reduce log level in production
- Use asynchronous logging
- Implement log sampling

**Missing context:**
- Ensure request IDs are passed through
- Add relevant fields to log calls
- Use structured logging consistently

### Debug Logging
Enable debug logging for troubleshooting:
```bash
# Set environment variable
export UMA_LOG_LEVEL=debug

# Or use command line flag
./uma boot --log-level=debug
```

## Best Practices

### Structured Fields
Always use structured fields instead of string formatting:
```go
// Good
logger.Info().
    Str("component", "api").
    Str("method", "GET").
    Int("status_code", 200).
    Msg("Request completed")

// Avoid
logger.Info().Msgf("Request completed: %s %d", "GET", 200)
```

### Request Context
Pass request context through the call stack:
```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := getRequestID(r)
    ctx := logger.WithContext(map[string]interface{}{
        "request_id": requestID,
        "component": "api",
    })
    
    processRequest(ctx, r)
}
```

### Error Logging
Include error context in error logs:
```go
logger.Error().
    Err(err).
    Str("component", "docker").
    Str("operation", "start").
    Str("container_id", containerID).
    Msg("Failed to start container")
```

## Next Steps

- **[Metrics Guide](metrics.md)** - Prometheus metrics documentation
- **[Testing Guide](testing.md)** - Running and writing tests
- **[Monitoring Setup](../deployment/monitoring.md)** - Complete observability stack
