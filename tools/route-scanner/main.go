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
)

// RouteInfo represents a discovered HTTP route
type RouteInfo struct {
	Path        string   `json:"path"`
	Methods     []string `json:"methods"`
	Handler     string   `json:"handler"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
}

// ScanResult represents the complete scan results
type ScanResult struct {
	TotalRoutes    int                    `json:"total_routes"`
	Categories     map[string]int         `json:"categories"`
	Routes         []RouteInfo            `json:"routes"`
	WebSocketPaths []string               `json:"websocket_paths"`
	PluginRoutes   []RouteInfo            `json:"plugin_routes"`
	Summary        map[string]interface{} `json:"summary"`
}

// RouteScanner scans the UMA codebase for HTTP routes
type RouteScanner struct {
	fileSet      *token.FileSet
	routes       []RouteInfo
	wsRoutes     []string
	pluginRoutes []RouteInfo
}

// NewRouteScanner creates a new route scanner
func NewRouteScanner() *RouteScanner {
	return &RouteScanner{
		fileSet:      token.NewFileSet(),
		routes:       make([]RouteInfo, 0),
		wsRoutes:     make([]string, 0),
		pluginRoutes: make([]RouteInfo, 0),
	}
}

// ScanDirectory scans a directory for Go files and extracts routes
func (rs *RouteScanner) ScanDirectory(dir string) error {
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

		return rs.scanFile(path)
	})
}

// scanFile scans a single Go file for route registrations
func (rs *RouteScanner) scanFile(filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse the Go file
	file, err := parser.ParseFile(rs.fileSet, filename, src, parser.ParseComments)
	if err != nil {
		return err
	}

	// Walk the AST to find route registrations
	ast.Inspect(file, func(n ast.Node) bool {
		rs.inspectNode(n, filename)
		return true
	})

	return nil
}

// inspectNode inspects an AST node for route patterns
func (rs *RouteScanner) inspectNode(n ast.Node, filename string) {
	switch node := n.(type) {
	case *ast.CallExpr:
		rs.handleCallExpr(node, filename)
	case *ast.GenDecl:
		rs.handleGenDecl(node, filename)
	}
}

// handleCallExpr handles function call expressions (like mux.HandleFunc)
func (rs *RouteScanner) handleCallExpr(call *ast.CallExpr, filename string) {
	// Look for HandleFunc calls (both mux.HandleFunc and r.mux.HandleFunc)
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "HandleFunc" {
			rs.extractRoute(call, filename)
		}
	}
}

// handleGenDecl handles general declarations (for WebSocket endpoint lists)
func (rs *RouteScanner) handleGenDecl(decl *ast.GenDecl, filename string) {
	if decl.Tok == token.VAR {
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for i, name := range valueSpec.Names {
					if strings.Contains(name.Name, "wsEndpoints") || strings.Contains(name.Name, "WebSocket") {
						if i < len(valueSpec.Values) {
							rs.extractWebSocketPaths(valueSpec.Values[i])
						}
					}
				}
			}
		}
	}
}

// extractRoute extracts route information from a HandleFunc call
func (rs *RouteScanner) extractRoute(call *ast.CallExpr, filename string) {
	if len(call.Args) < 2 {
		return
	}

	// Extract path (first argument)
	var path string
	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		path = strings.Trim(lit.Value, `"`)
	}

	// Extract handler (second argument)
	var handler string
	switch arg := call.Args[1].(type) {
	case *ast.SelectorExpr:
		if ident, ok := arg.X.(*ast.Ident); ok {
			handler = fmt.Sprintf("%s.%s", ident.Name, arg.Sel.Name)
		} else if sel, ok := arg.X.(*ast.SelectorExpr); ok {
			// Handle r.handler.Method pattern
			if ident, ok := sel.X.(*ast.Ident); ok {
				handler = fmt.Sprintf("%s.%s.%s", ident.Name, sel.Sel.Name, arg.Sel.Name)
			}
		}
	case *ast.Ident:
		handler = arg.Name
	}

	if path != "" && handler != "" {
		position := rs.fileSet.Position(call.Pos())

		route := RouteInfo{
			Path:        path,
			Methods:     rs.inferMethods(handler, path),
			Handler:     handler,
			File:        filename,
			Line:        position.Line,
			Category:    rs.categorizeRoute(path),
			Description: rs.generateDescription(path, handler),
		}

		// Determine if this is a plugin route
		if strings.Contains(filename, "plugins/") {
			rs.pluginRoutes = append(rs.pluginRoutes, route)
		} else {
			rs.routes = append(rs.routes, route)
		}
	}
}

// extractWebSocketPaths extracts WebSocket endpoint paths from array literals
func (rs *RouteScanner) extractWebSocketPaths(expr ast.Expr) {
	if comp, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range comp.Elts {
			if lit, ok := elt.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				path := strings.Trim(lit.Value, `"`)
				rs.wsRoutes = append(rs.wsRoutes, path)
			}
		}
	}
}

// inferMethods infers HTTP methods based on handler name and path patterns
func (rs *RouteScanner) inferMethods(handler, path string) []string {
	methods := []string{}

	// Default to GET for most handlers
	methods = append(methods, "GET")

	// Add POST for action endpoints
	if strings.Contains(handler, "Execute") ||
		strings.Contains(handler, "Start") ||
		strings.Contains(handler, "Stop") ||
		strings.Contains(handler, "Restart") ||
		strings.Contains(handler, "Reboot") ||
		strings.Contains(handler, "Shutdown") ||
		strings.Contains(handler, "Login") ||
		strings.Contains(handler, "Clear") ||
		strings.Contains(handler, "Mark") ||
		strings.Contains(handler, "Repair") ||
		strings.Contains(path, "/bulk/") ||
		strings.Contains(path, "/execute") ||
		strings.Contains(path, "/reboot") ||
		strings.Contains(path, "/shutdown") ||
		strings.Contains(path, "/login") ||
		strings.Contains(path, "/clear") ||
		strings.Contains(path, "/mark-all-read") ||
		strings.Contains(path, "/repair") {
		methods = append(methods, "POST")
	}

	// Add PUT/PATCH for update operations
	if strings.Contains(handler, "Update") || strings.Contains(handler, "Set") {
		methods = append(methods, "PUT")
	}

	// Add DELETE for delete operations
	if strings.Contains(handler, "Delete") || strings.Contains(handler, "Remove") {
		methods = append(methods, "DELETE")
	}

	return methods
}

// categorizeRoute categorizes a route based on its path
func (rs *RouteScanner) categorizeRoute(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/system"):
		return "System"
	case strings.HasPrefix(path, "/api/v1/storage"):
		return "Storage"
	case strings.HasPrefix(path, "/api/v1/docker"):
		return "Docker"
	case strings.HasPrefix(path, "/api/v1/vms"):
		return "VMs"
	case strings.HasPrefix(path, "/api/v1/auth"):
		return "Authentication"
	case strings.HasPrefix(path, "/api/v1/notifications"):
		return "Notifications"
	case strings.HasPrefix(path, "/api/v1/operations"):
		return "Operations"
	case strings.HasPrefix(path, "/api/v1/shares"):
		return "Shares"
	case strings.HasPrefix(path, "/api/v1/scripts"):
		return "Scripts"
	case strings.HasPrefix(path, "/api/v1/diagnostics"):
		return "Diagnostics"
	case strings.HasPrefix(path, "/api/v1/rate-limits"):
		return "Rate Limiting"
	case strings.HasPrefix(path, "/api/v1/ws"):
		return "WebSocket"
	case strings.HasPrefix(path, "/api/v1/health"):
		return "Health"
	case strings.HasPrefix(path, "/api/v1/docs"):
		return "Documentation"
	case strings.HasPrefix(path, "/metrics"):
		return "Metrics"
	default:
		return "Other"
	}
}

// generateDescription generates a human-readable description for a route
func (rs *RouteScanner) generateDescription(path, handler string) string {
	// Extract meaningful parts from handler name
	handlerParts := strings.Split(handler, ".")
	if len(handlerParts) > 1 {
		handlerName := handlerParts[1]

		// Remove "Handle" prefix if present
		if strings.HasPrefix(handlerName, "Handle") {
			handlerName = strings.TrimPrefix(handlerName, "Handle")
		}

		// Convert camelCase to readable format
		re := regexp.MustCompile(`([a-z])([A-Z])`)
		readable := re.ReplaceAllString(handlerName, `$1 $2`)

		return fmt.Sprintf("%s endpoint", readable)
	}

	return fmt.Sprintf("Endpoint for %s", path)
}

// GenerateReport generates a comprehensive scan report
func (rs *RouteScanner) GenerateReport() *ScanResult {
	// Sort routes by path for consistent output
	sort.Slice(rs.routes, func(i, j int) bool {
		return rs.routes[i].Path < rs.routes[j].Path
	})

	sort.Slice(rs.pluginRoutes, func(i, j int) bool {
		return rs.pluginRoutes[i].Path < rs.pluginRoutes[j].Path
	})

	// Count routes by category
	categories := make(map[string]int)
	for _, route := range rs.routes {
		categories[route.Category]++
	}

	// Generate summary
	summary := map[string]interface{}{
		"api_routes":      len(rs.routes),
		"plugin_routes":   len(rs.pluginRoutes),
		"websocket_paths": len(rs.wsRoutes),
		"total_endpoints": len(rs.routes) + len(rs.pluginRoutes) + len(rs.wsRoutes),
		"categories":      len(categories),
	}

	return &ScanResult{
		TotalRoutes:    len(rs.routes),
		Categories:     categories,
		Routes:         rs.routes,
		WebSocketPaths: rs.wsRoutes,
		PluginRoutes:   rs.pluginRoutes,
		Summary:        summary,
	}
}

// PrintReport prints a human-readable report to stdout
func (rs *RouteScanner) PrintReport() {
	result := rs.GenerateReport()

	fmt.Println("🔍 UMA Route Discovery Report")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total API Routes: %d\n", result.TotalRoutes)
	fmt.Printf("Plugin Routes: %d\n", len(result.PluginRoutes))
	fmt.Printf("WebSocket Paths: %d\n", len(result.WebSocketPaths))
	fmt.Printf("Total Endpoints: %d\n", result.Summary["total_endpoints"])
	fmt.Println()

	// Print routes by category
	fmt.Println("📊 Routes by Category:")
	for category, count := range result.Categories {
		fmt.Printf("  %s: %d\n", category, count)
	}
	fmt.Println()

	// Print all discovered routes
	fmt.Println("🛣️  Discovered Routes:")
	currentCategory := ""
	for _, route := range result.Routes {
		if route.Category != currentCategory {
			fmt.Printf("\n[%s]\n", route.Category)
			currentCategory = route.Category
		}

		methodsStr := strings.Join(route.Methods, ", ")
		fmt.Printf("  %-8s %-40s -> %s\n", methodsStr, route.Path, route.Handler)
	}

	// Print WebSocket paths
	if len(result.WebSocketPaths) > 0 {
		fmt.Println("\n🔌 WebSocket Endpoints:")
		for _, path := range result.WebSocketPaths {
			fmt.Printf("  WS       %-40s -> WebSocket Handler\n", path)
		}
	}

	// Print plugin routes
	if len(result.PluginRoutes) > 0 {
		fmt.Println("\n🔌 Plugin Routes:")
		for _, route := range result.PluginRoutes {
			methodsStr := strings.Join(route.Methods, ", ")
			fmt.Printf("  %-8s %-40s -> %s\n", methodsStr, route.Path, route.Handler)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: route-scanner <directory> [--json] [--output <file>]")
		fmt.Println("  directory: Root directory to scan (e.g., daemon/services/api)")
		fmt.Println("  --json: Output in JSON format")
		fmt.Println("  --output: Save output to file")
		os.Exit(1)
	}

	directory := os.Args[1]
	jsonOutput := false
	outputFile := ""

	// Parse command line arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--json":
			jsonOutput = true
		case "--output":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		}
	}

	// Create scanner and scan directory
	scanner := NewRouteScanner()

	if err := scanner.ScanDirectory(directory); err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	// Also scan plugins directory if it exists
	pluginsDir := "daemon/plugins"
	if _, err := os.Stat(pluginsDir); err == nil {
		if err := scanner.ScanDirectory(pluginsDir); err != nil {
			fmt.Printf("Warning: Error scanning plugins directory: %v\n", err)
		}
	}

	// Generate output
	if jsonOutput {
		result := scanner.GenerateReport()
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Error generating JSON: %v\n", err)
			os.Exit(1)
		}

		if outputFile != "" {
			err := os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Report saved to: %s\n", outputFile)
		} else {
			fmt.Println(string(jsonData))
		}
	} else {
		if outputFile != "" {
			// Redirect stdout to file for text output
			file, err := os.Create(outputFile)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()
			os.Stdout = file
		}

		scanner.PrintReport()

		if outputFile != "" {
			fmt.Fprintf(os.Stderr, "Report saved to: %s\n", outputFile)
		}
	}
}
