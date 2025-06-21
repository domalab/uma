package api

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"github.com/domalab/uma/daemon/common"
	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/plugins/diagnostics"
	"github.com/domalab/uma/daemon/plugins/docker"
	"github.com/domalab/uma/daemon/plugins/gpu"
	"github.com/domalab/uma/daemon/plugins/notifications"
	"github.com/domalab/uma/daemon/plugins/sensor"
	"github.com/domalab/uma/daemon/plugins/storage"
	"github.com/domalab/uma/daemon/plugins/system"
	"github.com/domalab/uma/daemon/plugins/ups"
	"github.com/domalab/uma/daemon/plugins/vm"
	"github.com/domalab/uma/daemon/services/api/events"
	"github.com/domalab/uma/daemon/services/async"
	"github.com/domalab/uma/daemon/services/auth"
	"github.com/domalab/uma/daemon/services/cache"
	"github.com/domalab/uma/daemon/services/config"
	"github.com/domalab/uma/daemon/services/mcp"
	upsDetector "github.com/domalab/uma/daemon/services/ups"
)

type Api struct {
	ctx *domain.Context

	// Unix socket listener
	listener net.Listener

	// HTTP server
	httpServer *HTTPServer

	// MCP server
	mcpServer *mcp.Server

	// Services
	configManager        *config.Manager
	authService          *auth.AuthService
	rateLimiter          *auth.RateLimiter
	operationRateLimiter *auth.OperationRateLimiter
	asyncManager         *async.AsyncManager
	eventManager         *events.EventManager

	// Data providers
	origin        *dto.Origin
	sensor        sensor.Sensor
	ups           ups.Ups
	upsDetector   *upsDetector.Detector
	storage       *storage.StorageMonitor
	system        *system.SystemMonitor
	gpu           *gpu.GPUMonitor
	docker        *docker.DockerManager
	vm            *vm.VMManager
	diagnostics   *diagnostics.DiagnosticsManager
	notifications *notifications.NotificationManager
}

func Create(ctx *domain.Context) *Api {
	// Initialize configuration manager
	configManager := config.NewManager("")
	if err := configManager.Load(); err != nil {
		logger.Yellow("Failed to load configuration: %v", err)
	}

	// Update context with loaded configuration
	loadedConfig := configManager.GetConfig()
	ctx.Config.Version = loadedConfig.Version
	if ctx.Config.Version == "" || ctx.Config.Version == "unknown" {
		ctx.Config.Version = loadedConfig.Version
	}

	// Initialize authentication service
	authService := auth.NewAuthService(loadedConfig.Auth)

	// Initialize rate limiter (100 requests per minute)
	rateLimiter := auth.NewRateLimiter(100, time.Minute)

	// Initialize operation-specific rate limiter
	operationRateLimiter := auth.NewOperationRateLimiter()

	// Initialize async manager
	asyncManager := async.NewAsyncManager()

	api := &Api{
		ctx:                  ctx,
		configManager:        configManager,
		authService:          authService,
		rateLimiter:          rateLimiter,
		operationRateLimiter: operationRateLimiter,
		asyncManager:         asyncManager,
	}

	// Initialize HTTP server if enabled
	if loadedConfig.HTTPServer.Enabled {
		api.httpServer = NewHTTPServer(api, loadedConfig.HTTPServer.Port)
	}

	// Initialize MCP server if enabled (requires HTTP server for API adapter)
	if loadedConfig.MCP.Enabled && api.httpServer != nil {
		// Create MCP configuration from domain config
		mcpConfig := config.MCPConfig{
			Enabled:        loadedConfig.MCP.Enabled,
			Port:           loadedConfig.MCP.Port,
			MaxConnections: loadedConfig.MCP.MaxConnections,
		}
		api.mcpServer = mcp.NewServer(mcpConfig, api.httpServer.apiAdapter)
	}

	return api
}

func (a *Api) Run() error {
	// Initialize all monitoring plugins
	a.sensor = a.createSensor()

	// Initialize UPS detector and start automatic detection
	a.upsDetector = upsDetector.NewDetector()

	// Add callback to refresh UPS instance when detection status changes
	a.upsDetector.AddStatusChangeCallback(func(available bool, upsType ups.Kind) {
		logger.Blue("UPS detection status changed, refreshing UPS instance...")
		a.RefreshUPS()
	})

	a.upsDetector.Start()
	a.ups = a.createUps()

	a.storage = storage.NewStorageMonitor()
	a.system = system.NewSystemMonitor()
	a.gpu = gpu.NewGPUMonitor()
	a.docker = docker.NewDockerManager()
	a.vm = vm.NewVMManager()
	a.diagnostics = diagnostics.NewDiagnosticsManager()
	a.notifications = notifications.NewNotificationManager()

	// Initialize cache system
	cache.InitializeGlobalInvalidator()

	// Register async operation executors
	a.registerAsyncExecutors()

	// Initialize and start event manager for enhanced WebSocket functionality
	if a.httpServer != nil && a.httpServer.enhancedWebSocketHandler != nil {
		a.eventManager = events.NewEventManager(a.httpServer.apiAdapter, a.ctx.Hub, a.httpServer.enhancedWebSocketHandler)
		a.eventManager.Start()
		logger.Green("Event Manager started for real-time monitoring")
	}

	// Start HTTP server if configured
	if a.httpServer != nil {
		if err := a.httpServer.Start(); err != nil {
			logger.Yellow("Failed to start HTTP server: %v", err)
		}
	}

	// Start MCP server if configured
	if a.mcpServer != nil {
		if err := a.mcpServer.Start(); err != nil {
			logger.Yellow("Failed to start MCP server: %v", err)
		}
	}

	// Start Unix socket server
	go a.startUnixSocketServer()

	return nil
}

// startUnixSocketServer starts the Unix socket API server
func (a *Api) startUnixSocketServer() {
	// make sure there's no socket file
	os.Remove(common.Socket)

	var err error
	a.listener, err = net.Listen("unix", common.Socket)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", common.Socket, err)
	}
	defer func() {
		a.listener.Close()
		os.Remove(common.Socket)
	}()

	logger.Blue("Unix socket API listening on %s", common.Socket)

	for {
		conn, err := a.listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go a.handleUnixSocketConnection(conn)
	}
}

// Stop gracefully stops all API servers
func (a *Api) Stop() error {
	logger.Blue("Stopping API services...")

	// Stop UPS detector
	if a.upsDetector != nil {
		a.upsDetector.Stop()
	}

	// Stop HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Stop(); err != nil {
			logger.Yellow("Error stopping HTTP server: %v", err)
		}
	}

	// Stop MCP server
	if a.mcpServer != nil {
		if err := a.mcpServer.Stop(); err != nil {
			logger.Yellow("Error stopping MCP server: %v", err)
		}
	}

	// Stop event manager
	if a.eventManager != nil {
		a.eventManager.Stop()
	}

	// Stop async manager
	if a.asyncManager != nil {
		a.asyncManager.Stop()
	}

	// Stop cache system
	cache.GetGlobalCacheManager().Stop()

	// Stop Unix socket server
	if a.listener != nil {
		a.listener.Close()
		os.Remove(common.Socket)
	}

	return nil
}

// handleUnixSocketConnection handles Unix socket connections (legacy API)
func (a *Api) handleUnixSocketConnection(conn net.Conn) {
	defer conn.Close()

	var req dto.Request
	err := json.NewDecoder(conn).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %v", err)
		conn.Write([]byte(`{"error": "Invalid request"}` + "\n"))
		return
	}

	logger.LightGreen("received %+v ", req)

	var resp []byte
	switch req.Action {
	case "get_info":
		reply := a.getInfo()
		resp, _ = json.Marshal(reply)

	case "get_logs":
		params := req.Params
		logType := params["logType"]
		reply := a.getLogs(logType)
		resp, _ = json.Marshal(reply)

	case "get_origin":
		reply := a.getOrigin()
		resp, _ = json.Marshal(reply)

	default:
		resp, _ = json.Marshal(map[string]string{"error": "Unsupported action"})
	}

	logger.Yellow(" sending %+v", string(resp))

	conn.Write(resp)
	conn.Write([]byte("\n"))
}

func (a *Api) createSensor() sensor.Sensor {
	s, err := sensor.IdentifySensor()
	if err != nil {
		logger.Yellow("error identifying sensor: %s", err)
	} else {
		switch s {
		case sensor.SYSTEM:
			logger.Blue("created system sensor ...")
			return sensor.NewSystemSensor()
		case sensor.IPMI:
			logger.Blue("created ipmi sensor ...")
			return sensor.NewIpmiSensor()
		}
	}

	logger.Blue("no sensor detected ...")

	return sensor.NewNoSensor()
}

func (a *Api) createUps() ups.Ups {
	// Use automatic UPS detection instead of configuration flag
	if a.upsDetector != nil && a.upsDetector.IsAvailable() {
		upsType := a.upsDetector.GetUPSType()
		switch upsType {
		case ups.APC:
			logger.Blue("created apc ups (auto-detected)...")
			return ups.NewApc()
		case ups.NUT:
			logger.Blue("created nut ups (auto-detected)...")
			return ups.NewNut()
		}
	}

	logger.Blue("no ups detected or available...")
	return ups.NewNoUps()
}

// registerAsyncExecutors registers all async operation executors
func (a *Api) registerAsyncExecutors() {
	// Create adapters for existing services
	storageAdapter := async.NewStorageMonitorAdapter(a.storage)
	dockerAdapter := async.NewDockerManagerAdapter(a.docker)

	// Register parity check executor
	parityExecutor := async.NewParityCheckExecutor(storageAdapter)
	a.asyncManager.RegisterExecutor(parityExecutor)

	// Register array operation executors
	arrayStartExecutor := async.NewArrayStartExecutor(storageAdapter)
	a.asyncManager.RegisterExecutor(arrayStartExecutor)

	arrayStopExecutor := async.NewArrayStopExecutor(storageAdapter)
	a.asyncManager.RegisterExecutor(arrayStopExecutor)

	// Register SMART scan executor
	smartExecutor := async.NewSMARTScanExecutor(storageAdapter)
	a.asyncManager.RegisterExecutor(smartExecutor)

	// Register bulk container executor
	bulkContainerExecutor := async.NewBulkContainerExecutor(dockerAdapter)
	a.asyncManager.RegisterExecutor(bulkContainerExecutor)

	logger.Blue("Registered %d async operation executors", 5)
}

// GetDockerManager returns the Docker manager instance
func (a *Api) GetDockerManager() *docker.DockerManager {
	return a.docker
}

// GetStorageMonitor returns the storage monitor instance
func (a *Api) GetStorageMonitor() *storage.StorageMonitor {
	return a.storage
}

// GetSystemMonitor returns the system monitor instance
func (a *Api) GetSystemMonitor() *system.SystemMonitor {
	return a.system
}

// GetVMManager returns the VM manager instance
func (a *Api) GetVMManager() *vm.VMManager {
	return a.vm
}

// GetUPSDetector returns the UPS detector instance
func (a *Api) GetUPSDetector() *upsDetector.Detector {
	return a.upsDetector
}

// RefreshUPS recreates the UPS instance based on current detection status
func (a *Api) RefreshUPS() {
	a.ups = a.createUps()
}

// GetMCPServer returns the MCP server instance
func (a *Api) GetMCPServer() *mcp.Server {
	return a.mcpServer
}
