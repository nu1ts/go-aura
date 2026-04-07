package xray

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

// validateJSON checks JSON syntax validity.
// Does not validate xray config semantics (that's done by parseConfig).
func validateJSON(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("empty config")
	}
	var js json.RawMessage
	if err := json.Unmarshal([]byte(s), &js); err != nil {
		return fmt.Errorf("malformed JSON: %w", err)
	}
	return nil
}

// parseConfig converts JSON config to xray protobuf format.
// Uses official xray parser for structure validation.
func parseConfig(configJSON string) (*core.Config, error) {
	reader := strings.NewReader(configJSON)
	pbConfig, err := serial.LoadJSONConfig(reader)
	if err != nil {
		return nil, err
	}
	return pbConfig, nil
}

// inboundConfig represents a partial inbound configuration for port extraction.
type inboundConfig struct {
	Port     json.RawMessage `json:"port"`
	Listen   string          `json:"listen"`
	Protocol string          `json:"protocol"`
}

// partialConfig represents a partial xray config for port extraction.
type partialConfig struct {
	Inbounds []inboundConfig `json:"inbounds"`
}

// checkPorts extracts ports from inbounds and verifies availability.
// Implements fail-fast pattern to catch port conflicts early.
//
// Handles:
//   - Numeric ports: "port": 1080
//   - Skips strings and ranges: "port": "10800-10810"
//   - Default listen address: 127.0.0.1
func (c *Core) checkPorts(configJSON string) error {
	var cfg partialConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		// Non-fatal: primary validation happens in parseConfig
		return nil
	}

	for _, inbound := range cfg.Inbounds {
		if inbound.Port == nil {
			continue
		}

		var port int
		if err := json.Unmarshal(inbound.Port, &port); err != nil {
			// Port might be a string or range — skip
			continue
		}

		listen := inbound.Listen
		if listen == "" {
			listen = "127.0.0.1"
		}

		if err := checkPortAvailable(listen, port); err != nil {
			return &PortError{Port: port, Err: err}
		}
	}

	return nil
}

// checkPortAvailable attempts to bind a port to verify availability.
// Uses net.Listen() followed by immediate Close().
//
// Limitations:
//   - Race condition: port may be taken between Close() and xray start
//   - Only checks TCP ports
func checkPortAvailable(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d on %s is not available: %w", port, host, err)
	}
	_ = ln.Close()
	return nil
}

// isPortError determines if an error is related to port binding.
// Cross-platform heuristic (Linux/macOS/Windows).
func isPortError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "address already in use") ||
		strings.Contains(errStr, "bind: address already in use") ||
		strings.Contains(errStr, "only one usage of each socket address")
}

// extractPort attempts to parse port number from error message.
// TODO: implement regex parsing for port extraction.
func extractPort(_ error) int {
	return 0
}
