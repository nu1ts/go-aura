package xray

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all" // Register all xray protocols
)

const (
	defaultStartTimeout    = 10 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

// Core manages the xray-core instance lifecycle.
// All public methods are thread-safe.
//
// Zero value is not usable; create instances with [New].
type Core struct {
	mu       sync.RWMutex
	instance *core.Instance
	state    State

	lastConfigJSON string

	startTimeout    time.Duration
	shutdownTimeout time.Duration

	logger Logger
}

// New creates a new [Core] instance with the given options.
// The returned core is in [StateStopped] and ready to start.
//
// Example:
//
//	// With defaults
//	core := xray.New()
//
//	// With custom configuration
//	core := xray.New(
//	    xray.WithLogger(logger),
//	    xray.WithStartTimeout(15 * time.Second),
//	    xray.WithShutdownTimeout(3 * time.Second),
//	)
func New(opts ...Option) *Core {
	c := &Core{
		state:           StateStopped,
		startTimeout:    defaultStartTimeout,
		shutdownTimeout: defaultShutdownTimeout,
		logger:          &defaultLogger{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// State returns the current state of the xray-core.
// Thread-safe: uses read lock for concurrent access.
func (c *Core) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// IsRunning reports whether the xray-core is currently running.
// Equivalent to State() == StateRunning.
func (c *Core) IsRunning() bool {
	return c.State() == StateRunning
}

// Start launches xray-core with the provided JSON configuration.
//
// The startup process:
//  1. Validates JSON syntax
//  2. Checks port availability (fail-fast)
//  3. Parses config to xray protobuf format
//  4. Creates xray instance
//  5. Starts instance with timeout protection
//
// Start is idempotent: calling Start on a running core returns [ErrAlreadyRunning].
//
// Errors:
//   - [ErrAlreadyRunning]: core is already running or starting
//   - [*ConfigError]: invalid JSON or xray configuration
//   - [*PortError]: one or more ports are unavailable
//   - [ErrStartTimeout]: startup exceeded timeout
//
// Example:
//
//	config := `{
//	    "inbounds": [{"port": 1080, "protocol": "socks"}],
//	    "outbounds": [{"protocol": "freedom"}]
//	}`
//
//	if err := core.Start(config); err != nil {
//	    var portErr *xray.PortError
//	    if errors.As(err, &portErr) {
//	        log.Printf("Port %d is already in use", portErr.Port)
//	        return
//	    }
//	    log.Fatal(err)
//	}
func (c *Core) Start(configJSON string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateRunning || c.state == StateStarting {
		return ErrAlreadyRunning
	}

	c.state = StateStarting
	c.logger.Info("Starting xray core...")

	// Fail-fast: validate JSON syntax
	if err := validateJSON(configJSON); err != nil {
		c.state = StateStopped
		return &ConfigError{Reason: "invalid JSON", Err: err}
	}

	// Fail-fast: check port availability before starting
	if err := c.checkPorts(configJSON); err != nil {
		c.state = StateStopped
		return err
	}

	// Parse JSON to xray protobuf config
	config, err := parseConfig(configJSON)
	if err != nil {
		c.state = StateStopped
		return &ConfigError{Reason: "failed to parse xray config", Err: err}
	}

	// Create xray instance
	instance, err := core.New(config)
	if err != nil {
		c.state = StateStopped
		return &ConfigError{Reason: "failed to create xray instance", Err: err}
	}

	// Start with timeout protection
	if err := c.startWithTimeout(instance); err != nil {
		return err
	}

	c.instance = instance
	c.lastConfigJSON = configJSON
	c.state = StateRunning
	c.logger.Info("Xray core started successfully")

	return nil
}

// startWithTimeout starts the instance with timeout protection.
// Uses goroutine + select pattern to prevent blocking.
func (c *Core) startWithTimeout(instance *core.Instance) error {
	// Buffered channel prevents goroutine leak if timeout fires first
	startDone := make(chan error, 1)
	go func() {
		startDone <- instance.Start()
	}()

	select {
	case err := <-startDone:
		if err != nil {
			c.state = StateStopped
			if isPortError(err) {
				return &PortError{Port: extractPort(err), Err: err}
			}
			return fmt.Errorf("failed to start xray core: %w", err)
		}
		return nil

	case <-time.After(c.startTimeout):
		_ = instance.Close() // Attempt to clean up
		c.state = StateStopped
		return ErrStartTimeout
	}
}

// Stop performs graceful shutdown of xray-core with timeout protection.
//
// The shutdown process:
//  1. Validates current state (must be [StateRunning])
//  2. Transitions to [StateStopping]
//  3. Calls instance.Close() in a goroutine
//  4. Waits for completion or timeout
//  5. Transitions to [StateStopped]
//
// Stop is safe to call multiple times: subsequent calls return [ErrNotRunning].
//
// Errors:
//   - [ErrNotRunning]: core is not currently running
//   - [ErrShutdownTimeout]: graceful shutdown exceeded timeout
//   - Other errors from xray instance.Close()
//
// Example:
//
//	if err := core.Stop(); err != nil {
//	    if errors.Is(err, xray.ErrShutdownTimeout) {
//	        log.Warn("Forced shutdown due to timeout")
//	    }
//	}
func (c *Core) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.stopLocked()
}

// stopLocked performs shutdown without acquiring the mutex.
// Used by Restart which already holds the lock.
func (c *Core) stopLocked() error {
	if c.state != StateRunning {
		return ErrNotRunning
	}

	c.state = StateStopping
	c.logger.Info("Stopping xray core...")

	// Graceful shutdown with timeout
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- c.instance.Close()
	}()

	var err error
	select {
	case err = <-shutdownDone:
		if err != nil {
			c.logger.Error("Error during xray core shutdown: %v", err)
		}
	case <-time.After(c.shutdownTimeout):
		c.logger.Error("Xray core shutdown timeout, forcing stop")
		err = ErrShutdownTimeout
	}

	c.instance = nil
	c.state = StateStopped
	c.logger.Info("Xray core stopped")

	return err
}

// Restart stops the running core and starts it with a new or existing configuration.
//
// Behavior:
//   - Without arguments: uses the last successful configuration
//   - With argument: applies the new configuration
//
// Errors:
//   - [*ConfigError]: no configuration available
//   - All errors from Start
//
// Example:
//
//	// Restart with current config
//	if err := core.Restart(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Apply new configuration
//	if err := core.Restart(newConfig); err != nil {
//	    log.Fatal(err)
//	}
func (c *Core) Restart(configJSON ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newConfig := c.lastConfigJSON
	if len(configJSON) > 0 && configJSON[0] != "" {
		newConfig = configJSON[0]
	}

	if newConfig == "" {
		return &ConfigError{Reason: "no configuration available for restart"}
	}

	c.logger.Info("Restarting xray core...")

	// Stop if currently running
	if c.state == StateRunning {
		if err := c.stopLocked(); err != nil && !errors.Is(err, ErrNotRunning) {
			c.logger.Error("Error stopping during restart: %v", err)
		}
	}

	// Temporarily release lock since Start() acquires it
	c.mu.Unlock()
	err := c.Start(newConfig)
	c.mu.Lock() // Re-acquire for deferred Unlock

	if err != nil {
		return fmt.Errorf("restart failed: %w", err)
	}

	c.logger.Info("Xray core restarted successfully")
	return nil
}

// GetLastConfig returns the last successfully applied JSON configuration.
// Returns an empty string if Start was never called successfully.
// Thread-safe.
func (c *Core) GetLastConfig() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastConfigJSON
}
