package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/domalab/uma/daemon/services/api/services"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// ScriptHandler handles script-related HTTP requests
type ScriptHandler struct {
	scriptService *services.ScriptService
	apiAdapter    utils.APIInterface
}

// NewScriptHandler creates a new script handler instance
func NewScriptHandler(apiAdapter utils.APIInterface) *ScriptHandler {
	return &ScriptHandler{
		scriptService: services.NewScriptService(apiAdapter),
		apiAdapter:    apiAdapter,
	}
}

// HandleScriptsList handles GET /api/v1/scripts - lists all available scripts
func (h *ScriptHandler) HandleScriptsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	scripts, err := h.scriptService.GetUserScripts()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user scripts: %v", err))
		return
	}

	response := map[string]interface{}{
		"scripts": scripts,
		"count":   len(scripts),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// HandleScript handles script operations
func (h *ScriptHandler) HandleScript(w http.ResponseWriter, r *http.Request) {
	// Extract script name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/scripts/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.WriteError(w, http.StatusBadRequest, "Script name required")
		return
	}

	scriptName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		switch action {
		case "status":
			status, err := h.scriptService.GetScriptStatus(scriptName)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get script status: %v", err))
				return
			}
			utils.WriteJSON(w, http.StatusOK, status)

		case "logs":
			logs, err := h.scriptService.GetScriptLogs(scriptName)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get script logs: %v", err))
				return
			}
			response := responses.ScriptLogsResponse{Name: scriptName, Logs: logs}
			utils.WriteJSON(w, http.StatusOK, response)

		default:
			utils.WriteError(w, http.StatusBadRequest, "Invalid action. Use 'status' or 'logs'")
		}

	case http.MethodPost:
		switch action {
		case "execute":
			var req requests.ScriptExecuteRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
				return
			}

			// Convert to service request format
			serviceReq := services.ScriptExecuteRequest{
				Background: req.Background,
				Arguments:  make(map[string]string),
			}
			for _, arg := range req.Arguments {
				serviceReq.Arguments[arg] = arg // Simple conversion for now
			}

			response, err := h.scriptService.ExecuteScript(scriptName, serviceReq)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to execute script: %v", err))
				return
			}
			utils.WriteJSON(w, http.StatusOK, response)

		case "stop":
			err := h.scriptService.StopScript(scriptName)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop script: %v", err))
				return
			}
			utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Script stopped successfully"})

		default:
			utils.WriteError(w, http.StatusBadRequest, "Invalid action. Use 'execute' or 'stop'")
		}

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
