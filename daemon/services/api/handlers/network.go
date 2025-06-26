package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/services/api/utils"
)

// NetworkHandler handles network-related API endpoints
type NetworkHandler struct {
	api utils.APIInterface
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(api utils.APIInterface) *NetworkHandler {
	return &NetworkHandler{
		api: api,
	}
}

// HandleNetworkInterfaces handles GET /api/v1/network/interfaces
func (h *NetworkHandler) HandleNetworkInterfaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	interfaces, err := h.getEnhancedNetworkInterfaces()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get network interfaces: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"interfaces":   interfaces,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleNetworkStats handles GET /api/v1/network/stats
func (h *NetworkHandler) HandleNetworkStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.getNetworkStats()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get network stats: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, stats)
}

// HandleNetworkConnections handles GET /api/v1/network/connections
func (h *NetworkHandler) HandleNetworkConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	connections, err := h.getNetworkConnections()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get network connections: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, connections)
}

// HandleNetworkPing handles POST /api/v1/network/ping
func (h *NetworkHandler) HandleNetworkPing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body for host parameter
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	host, ok := request["host"].(string)
	if !ok || host == "" {
		utils.WriteError(w, http.StatusBadRequest, "Host parameter is required")
		return
	}

	result, err := h.pingHost(host)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to ping host: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, result)
}

// getEnhancedNetworkInterfaces gets detailed network interface information
func (h *NetworkHandler) getEnhancedNetworkInterfaces() ([]interface{}, error) {
	// Get basic network info from system
	networkInfo, err := h.api.GetSystem().GetNetworkInfo()
	if err != nil {
		return nil, err
	}

	interfaces := make([]interface{}, 0)

	// Extract interfaces from the network info
	if netMap, ok := networkInfo.(map[string]interface{}); ok {
		// Handle the case where interfaces is []map[string]interface{} instead of []interface{}
		if ifaceList, ok := netMap["interfaces"].([]map[string]interface{}); ok {
			for _, iface := range ifaceList {
				enhanced := h.enhanceInterfaceInfo(iface)
				interfaces = append(interfaces, enhanced)
			}
		} else if ifaceList, ok := netMap["interfaces"].([]interface{}); ok {
			// Fallback for []interface{} type
			for _, iface := range ifaceList {
				if ifaceMap, ok := iface.(map[string]interface{}); ok {
					enhanced := h.enhanceInterfaceInfo(ifaceMap)
					interfaces = append(interfaces, enhanced)
				}
			}
		}
	}

	return interfaces, nil
}

// enhanceInterfaceInfo adds additional network interface details
func (h *NetworkHandler) enhanceInterfaceInfo(iface map[string]interface{}) map[string]interface{} {
	enhanced := make(map[string]interface{})

	// Copy existing fields
	for key, value := range iface {
		enhanced[key] = value
	}

	ifaceName, ok := iface["name"].(string)
	if !ok {
		return enhanced
	}

	// Add interface type first (this is safe)
	enhanced["type"] = h.getInterfaceType(ifaceName)

	// Only enhance physical interfaces to avoid errors on virtual ones
	if h.isPhysicalInterface(ifaceName) {
		// Add MAC address, MTU, speed, duplex using ethtool and ip commands
		if macAddr := h.getMACAddress(ifaceName); macAddr != "" {
			enhanced["mac_address"] = macAddr
		}

		if mtu := h.getMTU(ifaceName); mtu > 0 {
			enhanced["mtu"] = mtu
		}

		if speed := h.getInterfaceSpeed(ifaceName); speed > 0 {
			enhanced["speed_mbps"] = speed
		}

		if duplex := h.getDuplex(ifaceName); duplex != "" {
			enhanced["duplex"] = duplex
		}

		// Add link state
		enhanced["link_detected"] = h.isLinkDetected(ifaceName)
	} else {
		// For virtual interfaces, set reasonable defaults
		enhanced["mtu"] = 1500
		enhanced["duplex"] = "unknown"
		enhanced["speed_mbps"] = 0
		enhanced["link_detected"] = enhanced["status"] == "up"
	}

	return enhanced
}

// getMACAddress gets the MAC address for an interface
func (h *NetworkHandler) getMACAddress(ifaceName string) string {
	cmd := exec.Command("ip", "link", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "link/ether") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

// getMTU gets the MTU for an interface
func (h *NetworkHandler) getMTU(ifaceName string) int {
	cmd := exec.Command("ip", "link", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return 1500 // Default MTU
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "mtu") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "mtu" && i+1 < len(fields) {
					if mtu, err := strconv.Atoi(fields[i+1]); err == nil {
						return mtu
					}
				}
			}
		}
	}
	return 1500 // Default MTU
}

// getInterfaceSpeed gets the speed for an interface using ethtool
func (h *NetworkHandler) getInterfaceSpeed(ifaceName string) int {
	cmd := exec.Command("ethtool", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Speed:") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasSuffix(field, "Mb/s") {
					speedStr := strings.TrimSuffix(field, "Mb/s")
					if speed, err := strconv.Atoi(speedStr); err == nil {
						return speed
					}
				}
			}
		}
	}
	return 0
}

// getDuplex gets the duplex setting for an interface
func (h *NetworkHandler) getDuplex(ifaceName string) string {
	cmd := exec.Command("ethtool", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Duplex:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return strings.ToLower(fields[1])
			}
		}
	}
	return "unknown"
}

// isLinkDetected checks if link is detected for an interface
func (h *NetworkHandler) isLinkDetected(ifaceName string) bool {
	cmd := exec.Command("ethtool", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Link detected:") {
			return strings.Contains(line, "yes")
		}
	}
	return false
}

// getInterfaceType determines the interface type
func (h *NetworkHandler) getInterfaceType(ifaceName string) string {
	if strings.HasPrefix(ifaceName, "eth") {
		return "ethernet"
	} else if strings.HasPrefix(ifaceName, "wlan") {
		return "wireless"
	} else if strings.HasPrefix(ifaceName, "br") {
		return "bridge"
	} else if strings.HasPrefix(ifaceName, "docker") {
		return "docker"
	} else if strings.HasPrefix(ifaceName, "veth") {
		return "virtual_ethernet"
	} else if strings.HasPrefix(ifaceName, "virbr") {
		return "virtual_bridge"
	} else if strings.HasPrefix(ifaceName, "tun") {
		return "tunnel"
	}
	return "unknown"
}

// isPhysicalInterface determines if an interface is physical (not virtual)
func (h *NetworkHandler) isPhysicalInterface(ifaceName string) bool {
	// Physical interfaces that support ethtool
	return strings.HasPrefix(ifaceName, "eth") ||
		strings.HasPrefix(ifaceName, "wlan") ||
		strings.HasPrefix(ifaceName, "en") // systemd naming
}

// getNetworkStats gets comprehensive network statistics
func (h *NetworkHandler) getNetworkStats() (map[string]interface{}, error) {
	interfaces, err := h.getEnhancedNetworkInterfaces()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"interfaces":   interfaces,
		"summary":      h.calculateNetworkSummary(interfaces),
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	return stats, nil
}

// calculateNetworkSummary calculates overall network statistics
func (h *NetworkHandler) calculateNetworkSummary(interfaces []interface{}) map[string]interface{} {
	var totalRx, totalTx uint64
	activeInterfaces := 0

	for _, iface := range interfaces {
		if ifaceMap, ok := iface.(map[string]interface{}); ok {
			if status, ok := ifaceMap["status"].(string); ok && status == "up" {
				activeInterfaces++
			}

			if rxBytes, ok := ifaceMap["rx_bytes"].(uint64); ok {
				totalRx += rxBytes
			}

			if txBytes, ok := ifaceMap["tx_bytes"].(uint64); ok {
				totalTx += txBytes
			}
		}
	}

	return map[string]interface{}{
		"total_interfaces":    len(interfaces),
		"active_interfaces":   activeInterfaces,
		"total_rx_bytes":      totalRx,
		"total_tx_bytes":      totalTx,
		"total_traffic_bytes": totalRx + totalTx,
	}
}

// getNetworkConnections gets active network connections
func (h *NetworkHandler) getNetworkConnections() (map[string]interface{}, error) {
	tcpConnections, err := h.getTCPConnections()
	if err != nil {
		return nil, err
	}

	udpConnections, err := h.getUDPConnections()
	if err != nil {
		return nil, err
	}

	listeningPorts, err := h.getListeningPorts()
	if err != nil {
		return nil, err
	}

	connections := map[string]interface{}{
		"tcp_connections": tcpConnections,
		"udp_connections": udpConnections,
		"listening_ports": listeningPorts,
		"summary": map[string]interface{}{
			"total_connections": len(tcpConnections) + len(udpConnections),
			"tcp_connections":   len(tcpConnections),
			"udp_connections":   len(udpConnections),
			"listening_ports":   len(listeningPorts),
		},
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}

	return connections, nil
}

// getTCPConnections gets TCP connections using ss command
func (h *NetworkHandler) getTCPConnections() ([]interface{}, error) {
	cmd := exec.Command("ss", "-t", "-n")
	output, err := cmd.Output()
	if err != nil {
		return []interface{}{}, nil // Return empty slice on error
	}

	connections := make([]interface{}, 0)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			connection := map[string]interface{}{
				"protocol":      "tcp",
				"state":         fields[0],
				"local_address": fields[3],
				"peer_address":  fields[4],
			}
			connections = append(connections, connection)
		}
	}

	return connections, nil
}

// getUDPConnections gets UDP connections using ss command
func (h *NetworkHandler) getUDPConnections() ([]interface{}, error) {
	cmd := exec.Command("ss", "-u", "-n")
	output, err := cmd.Output()
	if err != nil {
		return []interface{}{}, nil // Return empty slice on error
	}

	connections := make([]interface{}, 0)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			connection := map[string]interface{}{
				"protocol":      "udp",
				"state":         fields[0],
				"local_address": fields[3],
				"peer_address":  fields[4],
			}
			connections = append(connections, connection)
		}
	}

	return connections, nil
}

// getListeningPorts gets listening ports using ss command
func (h *NetworkHandler) getListeningPorts() ([]interface{}, error) {
	ports := make([]interface{}, 0)

	// Get TCP listening ports
	tcpCmd := exec.Command("ss", "-tln")
	tcpOutput, err := tcpCmd.Output()
	if err == nil {
		lines := strings.Split(string(tcpOutput), "\n")
		for _, line := range lines[1:] { // Skip header
			if strings.TrimSpace(line) == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 5 && fields[0] == "LISTEN" {
				port := map[string]interface{}{
					"protocol":      "tcp",
					"local_address": fields[3],
					"state":         "listening",
				}
				ports = append(ports, port)
			}
		}
	}

	// Get UDP listening ports (UNCONN state)
	udpCmd := exec.Command("ss", "-uln")
	udpOutput, err := udpCmd.Output()
	if err == nil {
		lines := strings.Split(string(udpOutput), "\n")
		for _, line := range lines[1:] { // Skip header
			if strings.TrimSpace(line) == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 5 && fields[0] == "UNCONN" {
				port := map[string]interface{}{
					"protocol":      "udp",
					"local_address": fields[3],
					"state":         "listening",
				}
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

// pingHost performs a network connectivity test
func (h *NetworkHandler) pingHost(host string) (map[string]interface{}, error) {
	cmd := exec.Command("ping", "-c", "4", "-W", "3", host)
	output, err := cmd.Output()

	result := map[string]interface{}{
		"host":      host,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		result["reachable"] = false
		result["error"] = err.Error()
		return result, nil
	}

	// Parse ping output
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	result["reachable"] = true
	result["packets_sent"] = 4
	result["packets_received"] = 4
	result["packet_loss"] = "0%"

	// Extract response time from last line
	for _, line := range lines {
		if strings.Contains(line, "min/avg/max") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Contains(field, "/") && strings.Contains(field, ".") {
					times := strings.Split(field, "/")
					if len(times) >= 2 {
						result["response_time"] = times[1] + "ms" // avg time
					}
					break
				}
			}
		}
		if strings.Contains(line, "packet loss") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasSuffix(field, "%") {
					result["packet_loss"] = field
					break
				}
			}
		}
	}

	return result, nil
}
