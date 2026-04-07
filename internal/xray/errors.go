package xray

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure conditions.
// Use [errors.Is] to check for these errors.
var (
	// ErrAlreadyRunning is returned when Start is called on a running core.
	ErrAlreadyRunning = errors.New("xray core is already running")

	// ErrNotRunning is returned when Stop is called on a stopped core.
	ErrNotRunning = errors.New("xray core is not running")

	// ErrInvalidConfig is a sentinel for configuration errors.
	// Use [errors.Is] with [*ConfigError] for detailed information.
	ErrInvalidConfig = errors.New("invalid xray configuration")

	// ErrPortInUse is a sentinel for port availability errors.
	// Use [errors.Is] with [*PortError] for detailed information.
	ErrPortInUse = errors.New("port is already in use")

	// ErrStartTimeout is returned when xray-core fails to start within timeout.
	ErrStartTimeout = errors.New("xray core start timeout")

	// ErrShutdownTimeout is returned when graceful shutdown exceeds timeout.
	ErrShutdownTimeout = errors.New("xray core shutdown timeout")
)

// ConfigError represents a configuration-related error with details.
//
// Example:
//
//	var cfgErr *xray.ConfigError
//	if errors.As(err, &cfgErr) {
//	    fmt.Printf("Config error: %s\n", cfgErr.Reason)
//	}
type ConfigError struct {
	Reason string // Human-readable description
	Err    error  // Underlying error (may be nil)
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error: %s: %v", e.Reason, e.Err)
	}
	return fmt.Sprintf("config error: %s", e.Reason)
}

// Unwrap returns the underlying error for [errors.Is] and [errors.As].
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// Is allows [errors.Is](err, ErrInvalidConfig) to work with ConfigError.
func (e *ConfigError) Is(target error) bool {
	return target == ErrInvalidConfig
}

// PortError represents a port availability error with port number.
//
// Example:
//
//	var portErr *xray.PortError
//	if errors.As(err, &portErr) {
//	    fmt.Printf("Port %d is busy\n", portErr.Port)
//	}
type PortError struct {
	Port int   // The unavailable port number
	Err  error // Underlying error
}

func (e *PortError) Error() string {
	if e.Port > 0 {
		return fmt.Sprintf("port %d is already in use: %v", e.Port, e.Err)
	}
	return fmt.Sprintf("port is already in use: %v", e.Err)
}

// Unwrap returns the underlying error for [errors.Is] and [errors.As].
func (e *PortError) Unwrap() error {
	return e.Err
}

// Is allows [errors.Is](err, ErrPortInUse) to work with PortError.
func (e *PortError) Is(target error) bool {
	return target == ErrPortInUse
}
