package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// WebSocketEndpoint represents a WebSocket endpoint
type WebSocketEndpoint struct {
	Path        string                 `json:"path"`
	Description string                 `json:"description"`
	Channels    []WebSocketChannel     `json:"channels"`
	Messages    []WebSocketMessage     `json:"messages"`
	Events      []WebSocketEvent       `json:"events"`
	Examples    map[string]interface{} `json:"examples"`
}

// WebSocketChannel represents a subscription channel
type WebSocketChannel struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EventTypes  []string `json:"event_types"`
	DataFormat  string   `json:"data_format"`
}

// WebSocketMessage represents a WebSocket message type
type WebSocketMessage struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Example     interface{}            `json:"example"`
}

// WebSocketEvent represents a WebSocket event
type WebSocketEvent struct {
	Name        string                 `json:"name"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	DataSchema  map[string]interface{} `json:"data_schema"`
	Example     interface{}            `json:"example"`
}

// WebSocketDocumentation represents complete WebSocket documentation
type WebSocketDocumentation struct {
	Endpoints    []WebSocketEndpoint    `json:"endpoints"`
	Channels     []WebSocketChannel     `json:"channels"`
	MessageTypes []WebSocketMessage     `json:"message_types"`
	EventTypes   []WebSocketEvent       `json:"event_types"`
	Summary      map[string]interface{} `json:"summary"`
	Generated    map[string]interface{} `json:"x-generated"`
}

// WebSocketDocumenter scans and documents WebSocket endpoints
type WebSocketDocumenter struct {
	fileSet   *token.FileSet
	endpoints []WebSocketEndpoint
	channels  []WebSocketChannel
	messages  []WebSocketMessage
	events    []WebSocketEvent
}

// NewWebSocketDocumenter creates a new WebSocket documenter
func NewWebSocketDocumenter() *WebSocketDocumenter {
	return &WebSocketDocumenter{
		fileSet:   token.NewFileSet(),
		endpoints: make([]WebSocketEndpoint, 0),
		channels:  make([]WebSocketChannel, 0),
		messages:  make([]WebSocketMessage, 0),
		events:    make([]WebSocketEvent, 0),
	}
}

// ScanDirectory scans a directory for WebSocket-related code
func (wd *WebSocketDocumenter) ScanDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		return wd.scanFile(path)
	})
}

// scanFile scans a single Go file for WebSocket patterns
func (wd *WebSocketDocumenter) scanFile(filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse the Go file
	file, err := parser.ParseFile(wd.fileSet, filename, src, parser.ParseComments)
	if err != nil {
		return err
	}

	// Walk the AST to find WebSocket patterns
	ast.Inspect(file, func(n ast.Node) bool {
		wd.inspectNode(n, filename)
		return true
	})

	return nil
}

// inspectNode inspects an AST node for WebSocket patterns
func (wd *WebSocketDocumenter) inspectNode(n ast.Node, filename string) {
	switch node := n.(type) {
	case *ast.GenDecl:
		wd.handleGenDecl(node, filename)
	case *ast.FuncDecl:
		wd.handleFuncDecl(node, filename)
	}
}

// handleGenDecl handles general declarations (constants, variables)
func (wd *WebSocketDocumenter) handleGenDecl(decl *ast.GenDecl, filename string) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			wd.handleValueSpec(s, filename)
		case *ast.TypeSpec:
			wd.handleTypeSpec(s, filename)
		}
	}
}

// handleValueSpec handles value specifications (constants, variables)
func (wd *WebSocketDocumenter) handleValueSpec(spec *ast.ValueSpec, filename string) {
	for _, name := range spec.Names {
		if strings.Contains(name.Name, "Event") && strings.Contains(filename, "websocket") {
			// Extract event type definitions
			wd.extractEventType(name.Name, spec, filename)
		}
	}
}

// handleTypeSpec handles type specifications
func (wd *WebSocketDocumenter) handleTypeSpec(spec *ast.TypeSpec, filename string) {
	if strings.Contains(spec.Name.Name, "WebSocket") || strings.Contains(spec.Name.Name, "Event") {
		// Extract WebSocket-related type definitions
		wd.extractWebSocketType(spec, filename)
	}
}

// handleFuncDecl handles function declarations
func (wd *WebSocketDocumenter) handleFuncDecl(decl *ast.FuncDecl, filename string) {
	if decl.Name != nil && strings.Contains(decl.Name.Name, "WebSocket") {
		// Extract WebSocket handler functions
		wd.extractWebSocketHandler(decl, filename)
	}
}

// extractEventType extracts event type information
func (wd *WebSocketDocumenter) extractEventType(name string, spec *ast.ValueSpec, filename string) {
	event := WebSocketEvent{
		Name:        name,
		Source:      wd.inferEventSource(name),
		Description: wd.generateEventDescription(name),
		DataSchema:  wd.generateEventSchema(name),
		Example:     wd.generateEventExample(name),
	}

	wd.events = append(wd.events, event)
}

// extractWebSocketType extracts WebSocket type information
func (wd *WebSocketDocumenter) extractWebSocketType(spec *ast.TypeSpec, filename string) {
	message := WebSocketMessage{
		Type:        spec.Name.Name,
		Description: wd.generateTypeDescription(spec.Name.Name),
		Schema:      wd.generateTypeSchema(spec.Name.Name),
		Example:     wd.generateTypeExample(spec.Name.Name),
	}

	wd.messages = append(wd.messages, message)
}

// extractWebSocketHandler extracts WebSocket handler information
func (wd *WebSocketDocumenter) extractWebSocketHandler(decl *ast.FuncDecl, filename string) {
	endpoint := WebSocketEndpoint{
		Path:        wd.inferEndpointPath(decl.Name.Name),
		Description: wd.generateHandlerDescription(decl.Name.Name),
		Channels:    wd.getDefaultChannels(),
		Messages:    wd.messages,
		Events:      wd.events,
		Examples:    wd.generateEndpointExamples(),
	}

	wd.endpoints = append(wd.endpoints, endpoint)
}

// Helper functions for generating documentation content
func (wd *WebSocketDocumenter) inferEventSource(eventName string) string {
	eventName = strings.ToLower(eventName)
	switch {
	case strings.Contains(eventName, "system"):
		return "system"
	case strings.Contains(eventName, "docker"):
		return "docker"
	case strings.Contains(eventName, "storage"):
		return "storage"
	case strings.Contains(eventName, "vm"):
		return "vm"
	case strings.Contains(eventName, "ups"):
		return "ups"
	case strings.Contains(eventName, "temperature"):
		return "monitoring"
	case strings.Contains(eventName, "disk"):
		return "storage"
	case strings.Contains(eventName, "network"):
		return "system"
	default:
		return "unknown"
	}
}

func (wd *WebSocketDocumenter) generateEventDescription(eventName string) string {
	// Convert camelCase to readable format
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	readable := re.ReplaceAllString(eventName, `$1 $2`)
	return fmt.Sprintf("Event for %s", strings.ToLower(readable))
}

func (wd *WebSocketDocumenter) generateEventSchema(eventName string) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"event_type": map[string]interface{}{
				"type":    "string",
				"example": strings.ToLower(eventName),
			},
			"data": map[string]interface{}{
				"type":        "object",
				"description": "Event-specific data payload",
			},
			"timestamp": map[string]interface{}{
				"type":   "string",
				"format": "date-time",
			},
		},
		"required": []string{"event_type", "data", "timestamp"},
	}
}

func (wd *WebSocketDocumenter) generateEventExample(eventName string) interface{} {
	return map[string]interface{}{
		"event_type": strings.ToLower(eventName),
		"data": map[string]interface{}{
			"message": "Sample event data",
			"value":   42,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
}

func (wd *WebSocketDocumenter) generateTypeDescription(typeName string) string {
	return fmt.Sprintf("WebSocket message type: %s", typeName)
}

func (wd *WebSocketDocumenter) generateTypeSchema(typeName string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": fmt.Sprintf("Schema for %s", typeName),
	}
}

func (wd *WebSocketDocumenter) generateTypeExample(typeName string) interface{} {
	return map[string]interface{}{
		"type": strings.ToLower(typeName),
		"data": map[string]interface{}{},
	}
}

func (wd *WebSocketDocumenter) inferEndpointPath(handlerName string) string {
	if strings.Contains(handlerName, "Unified") {
		return "/api/v1/ws"
	}
	return "/api/v1/ws/unknown"
}

func (wd *WebSocketDocumenter) generateHandlerDescription(handlerName string) string {
	return fmt.Sprintf("WebSocket handler: %s", handlerName)
}

func (wd *WebSocketDocumenter) getDefaultChannels() []WebSocketChannel {
	return []WebSocketChannel{
		{
			Name:        "system.stats",
			Description: "Real-time system performance metrics",
			EventTypes:  []string{"system.stats", "cpu.stats", "memory.stats"},
			DataFormat:  "JSON",
		},
		{
			Name:        "docker.events",
			Description: "Docker container lifecycle events",
			EventTypes:  []string{"docker.container.start", "docker.container.stop"},
			DataFormat:  "JSON",
		},
		{
			Name:        "storage.status",
			Description: "Storage array and disk status updates",
			EventTypes:  []string{"storage.array.status", "disk.smart.warning"},
			DataFormat:  "JSON",
		},
	}
}

func (wd *WebSocketDocumenter) generateEndpointExamples() map[string]interface{} {
	return map[string]interface{}{
		"subscribe": map[string]interface{}{
			"type":    "subscribe",
			"channel": "system.stats",
		},
		"unsubscribe": map[string]interface{}{
			"type":    "unsubscribe",
			"channel": "system.stats",
		},
		"ping": map[string]interface{}{
			"type": "ping",
		},
	}
}

// GenerateDocumentation generates complete WebSocket documentation
func (wd *WebSocketDocumenter) GenerateDocumentation() *WebSocketDocumentation {
	// Sort for consistent output
	sort.Slice(wd.endpoints, func(i, j int) bool {
		return wd.endpoints[i].Path < wd.endpoints[j].Path
	})

	sort.Slice(wd.events, func(i, j int) bool {
		return wd.events[i].Name < wd.events[j].Name
	})

	// Generate summary
	summary := map[string]interface{}{
		"total_endpoints":     len(wd.endpoints),
		"total_channels":      len(wd.channels),
		"total_message_types": len(wd.messages),
		"total_event_types":   len(wd.events),
	}

	return &WebSocketDocumentation{
		Endpoints:    wd.endpoints,
		Channels:     wd.getDefaultChannels(), // Use predefined channels
		MessageTypes: wd.messages,
		EventTypes:   wd.events,
		Summary:      summary,
		Generated: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"generator": "UMA WebSocket Documenter",
			"version":   "1.0.0",
		},
	}
}

// LoadExistingSchemas loads existing WebSocket schemas from the codebase
func (wd *WebSocketDocumenter) LoadExistingSchemas() error {
	// Load predefined channels from UMA documentation
	wd.channels = []WebSocketChannel{
		{
			Name:        "system.stats",
			Description: "Real-time system performance metrics",
			EventTypes:  []string{"system.stats", "cpu.stats", "memory.stats", "network.stats"},
			DataFormat:  "JSON",
		},
		{
			Name:        "docker.events",
			Description: "Docker container lifecycle events",
			EventTypes:  []string{"docker.container.start", "docker.container.stop", "docker.container.restart"},
			DataFormat:  "JSON",
		},
		{
			Name:        "storage.status",
			Description: "Storage array and disk status updates",
			EventTypes:  []string{"storage.array.status", "disk.smart.warning", "parity.status"},
			DataFormat:  "JSON",
		},
		{
			Name:        "temperature.alert",
			Description: "Temperature threshold alerts",
			EventTypes:  []string{"temperature.alert", "temperature.warning"},
			DataFormat:  "JSON",
		},
		{
			Name:        "resource.alert",
			Description: "CPU/memory/disk usage alerts",
			EventTypes:  []string{"resource.alert", "cpu.alert", "memory.alert"},
			DataFormat:  "JSON",
		},
		{
			Name:        "infrastructure.status",
			Description: "UPS, fans, power monitoring",
			EventTypes:  []string{"ups.status", "fan.status", "power.status"},
			DataFormat:  "JSON",
		},
	}

	// Load predefined message types
	wd.messages = []WebSocketMessage{
		{
			Type:        "subscribe",
			Description: "Subscribe to a WebSocket channel",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"subscribe"},
					},
					"channel": map[string]interface{}{
						"type": "string",
						"enum": []string{"system.stats", "docker.events", "storage.status", "temperature.alert", "resource.alert", "infrastructure.status"},
					},
				},
				"required": []string{"type", "channel"},
			},
			Example: map[string]interface{}{
				"type":    "subscribe",
				"channel": "system.stats",
			},
		},
		{
			Type:        "unsubscribe",
			Description: "Unsubscribe from a WebSocket channel",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"unsubscribe"},
					},
					"channel": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"type", "channel"},
			},
			Example: map[string]interface{}{
				"type":    "unsubscribe",
				"channel": "system.stats",
			},
		},
		{
			Type:        "event",
			Description: "WebSocket event message",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"event"},
					},
					"event_type": map[string]interface{}{
						"type": "string",
					},
					"data": map[string]interface{}{
						"type": "object",
					},
					"timestamp": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
				"required": []string{"type", "event_type", "data", "timestamp"},
			},
			Example: map[string]interface{}{
				"type":       "event",
				"event_type": "system.stats",
				"data": map[string]interface{}{
					"cpu_percent":    25.5,
					"memory_percent": 45.2,
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	// Add main WebSocket endpoint
	wd.endpoints = []WebSocketEndpoint{
		{
			Path:        "/api/v1/ws",
			Description: "Unified WebSocket endpoint with subscription management for real-time monitoring",
			Channels:    wd.channels,
			Messages:    wd.messages,
			Events:      wd.events,
			Examples:    wd.generateEndpointExamples(),
		},
	}

	return nil
}

// SaveDocumentation saves the WebSocket documentation to a file
func (wd *WebSocketDocumenter) SaveDocumentation(filename string) error {
	doc := wd.GenerateDocumentation()

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal documentation: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// PrintSummary prints a summary of the WebSocket documentation
func (wd *WebSocketDocumenter) PrintSummary() {
	doc := wd.GenerateDocumentation()

	fmt.Printf("🔌 WebSocket Documentation Summary:\n")
	fmt.Printf("  Endpoints: %d\n", len(doc.Endpoints))
	fmt.Printf("  Channels: %d\n", len(doc.Channels))
	fmt.Printf("  Message Types: %d\n", len(doc.MessageTypes))
	fmt.Printf("  Event Types: %d\n", len(doc.EventTypes))

	fmt.Printf("\n📡 Available Channels:\n")
	for _, channel := range doc.Channels {
		fmt.Printf("  %s: %s\n", channel.Name, channel.Description)
	}

	fmt.Printf("\n💬 Message Types:\n")
	for _, msg := range doc.MessageTypes {
		fmt.Printf("  %s: %s\n", msg.Type, msg.Description)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: websocket-documenter <directory> [--output <file>]")
		fmt.Println("  directory: Root directory to scan (e.g., daemon/services/api)")
		fmt.Println("  --output: Output file for WebSocket documentation (default: websocket-docs.json)")
		os.Exit(1)
	}

	directory := os.Args[1]
	outputFile := "websocket-docs.json"

	// Parse command line arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--output":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		}
	}

	documenter := NewWebSocketDocumenter()

	fmt.Printf("🔍 Scanning WebSocket patterns in: %s\n", directory)

	// Load existing schemas first
	if err := documenter.LoadExistingSchemas(); err != nil {
		fmt.Printf("Warning: Could not load existing schemas: %v\n", err)
	}

	// Scan directory for additional patterns
	if err := documenter.ScanDirectory(directory); err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	// Save documentation
	if err := documenter.SaveDocumentation(outputFile); err != nil {
		fmt.Printf("Error saving documentation: %v\n", err)
		os.Exit(1)
	}

	documenter.PrintSummary()
	fmt.Printf("\n✅ WebSocket documentation saved to: %s\n", outputFile)
}
