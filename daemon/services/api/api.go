package api

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"github.com/domalab/omniraid/daemon/common"
	"github.com/domalab/omniraid/daemon/domain"
	"github.com/domalab/omniraid/daemon/dto"
	"github.com/domalab/omniraid/daemon/logger"
	"github.com/domalab/omniraid/daemon/plugins/sensor"
	"github.com/domalab/omniraid/daemon/plugins/ups"
	"github.com/domalab/omniraid/daemon/plugins/storage"
	"github.com/domalab/omniraid/daemon/plugins/system"
	"github.com/domalab/omniraid/daemon/plugins/gpu"
	"github.com/domalab/omniraid/daemon/plugins/docker"
	"github.com/domalab/omniraid/daemon/plugins/vm"
	"github.com/domalab/omniraid/daemon/plugins/diagnostics"
	"github.com/domalab/omniraid/daemon/services/auth"
	"github.com/domalab/omniraid/daemon/services/config"
)

type Api struct {
	ctx *domain.Context

	// Unix socket listener
	listener net.Listener

	// HTTP server
	httpServer *HTTPServer

	// Services
	configManager *config.Manager
	authService   *auth.AuthService
	rateLimiter   *auth.RateLimiter

	// Data providers
	origin      *dto.Origin
	sensor      sensor.Sensor
	ups         ups.Ups
	storage     *storage.StorageMonitor
	system      *system.SystemMonitor
	gpu         *gpu.GPUMonitor
	docker      *docker.DockerManager
	vm          *vm.VMManager
	diagnostics *diagnostics.DiagnosticsManager
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

	api := &Api{
		ctx:           ctx,
		configManager: configManager,
		authService:   authService,
		rateLimiter:   rateLimiter,
	}

	// Initialize HTTP server if enabled
	if loadedConfig.HTTPServer.Enabled {
		api.httpServer = NewHTTPServer(api, loadedConfig.HTTPServer.Port)
	}

	return api
}

func (a *Api) Run() error {
	// Initialize all monitoring plugins
	a.sensor = a.createSensor()
	a.ups = a.createUps()
	a.storage = storage.NewStorageMonitor()
	a.system = system.NewSystemMonitor()
	a.gpu = gpu.NewGPUMonitor()
	a.docker = docker.NewDockerManager()
	a.vm = vm.NewVMManager()
	a.diagnostics = diagnostics.NewDiagnosticsManager()

	// Start HTTP server if configured
	if a.httpServer != nil {
		if err := a.httpServer.Start(); err != nil {
			logger.Yellow("Failed to start HTTP server: %v", err)
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

	// Stop HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Stop(); err != nil {
			logger.Yellow("Error stopping HTTP server: %v", err)
		}
	}

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
	logger.Blue("showing ups %t ...", a.ctx.Config.ShowUps)
	if a.ctx.Config.ShowUps {
		u, err := ups.IdentifyUps()
		if err != nil {
			logger.Yellow("error identifying ups: %s", err)
		} else {
			switch u {
			case ups.APC:
				logger.Blue("created apc ups ...")
				return ups.NewApc()
			case ups.NUT:
				logger.Blue("created nut ups ...")
				return ups.NewNut()
			}
		}
	}

	logger.Blue("no ups detected ...")

	return ups.NewNoUps()
}
