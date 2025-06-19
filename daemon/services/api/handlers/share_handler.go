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

// ShareHandler handles share-related HTTP requests
type ShareHandler struct {
	shareService *services.ShareService
	apiAdapter   utils.APIInterface
}

// NewShareHandler creates a new share handler instance
func NewShareHandler(apiAdapter utils.APIInterface) *ShareHandler {
	return &ShareHandler{
		shareService: services.NewShareService(),
		apiAdapter:   apiAdapter,
	}
}

// HandleShares handles GET /api/v1/shares and POST /api/v1/shares
func (h *ShareHandler) HandleShares(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		shares, err := h.shareService.GetShares()
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get shares: %v", err))
			return
		}

		response := responses.ShareListResponse{Shares: shares}
		utils.WriteJSON(w, http.StatusOK, response)
		return
	}

	if r.Method == http.MethodPost {
		var req requests.ShareCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		if err := h.shareService.ValidateShareCreateRequest(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
			return
		}

		if err := h.shareService.CreateShare(&req); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create share: %v", err))
			return
		}

		response := responses.ShareOperationResponse{
			Success:   true,
			Message:   fmt.Sprintf("Share '%s' created successfully", req.Name),
			ShareName: req.Name,
		}
		utils.WriteJSON(w, http.StatusCreated, response)
		return
	}

	utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// HandleShare handles share operations for specific shares
func (h *ShareHandler) HandleShare(w http.ResponseWriter, r *http.Request) {
	// Extract share name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shares/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.WriteError(w, http.StatusBadRequest, "Share name required")
		return
	}

	shareName := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		if action == "usage" {
			usage, err := h.shareService.GetShareUsage(shareName)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get share usage: %v", err))
				return
			}
			utils.WriteJSON(w, http.StatusOK, usage)
		} else if action == "" {
			share, err := h.shareService.GetShare(shareName)
			if err != nil {
				utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("Share not found: %v", err))
				return
			}
			utils.WriteJSON(w, http.StatusOK, share)
		} else {
			utils.WriteError(w, http.StatusBadRequest, "Invalid action. Use 'usage' or omit for share details")
		}

	case http.MethodPut:
		if action != "" {
			utils.WriteError(w, http.StatusBadRequest, "Action not allowed for PUT requests")
			return
		}

		var req requests.ShareUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
			return
		}

		if err := h.shareService.ValidateShareUpdateRequest(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
			return
		}

		if err := h.shareService.UpdateShare(shareName, &req); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update share: %v", err))
			return
		}

		response := responses.ShareOperationResponse{
			Success:   true,
			Message:   fmt.Sprintf("Share '%s' updated successfully", shareName),
			ShareName: shareName,
		}
		utils.WriteJSON(w, http.StatusOK, response)

	case http.MethodDelete:
		if action != "" {
			utils.WriteError(w, http.StatusBadRequest, "Action not allowed for DELETE requests")
			return
		}

		if err := h.shareService.DeleteShare(shareName); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete share: %v", err))
			return
		}

		response := responses.ShareOperationResponse{
			Success:   true,
			Message:   fmt.Sprintf("Share '%s' deleted successfully", shareName),
			ShareName: shareName,
		}
		utils.WriteJSON(w, http.StatusOK, response)

	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
