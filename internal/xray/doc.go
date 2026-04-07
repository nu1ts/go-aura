// Package xray provides a thread-safe wrapper for managing xray-core lifecycle
// with graceful shutdown, configurable timeouts, and comprehensive error handling.
//
// # Architecture
//
// The package uses several design patterns:
//   - FSM (Finite State Machine) for state management
//   - RWMutex for optimized concurrent reads
//   - Timeout pattern for non-blocking operations
//   - Functional options for configuration
//   - Dependency injection for logging
//
// # Basic Usage
//
//	core := xray.New()
//	if err := core.Start(configJSON); err != nil {
//	    log.Fatal(err)
//	}
//	defer core.Stop()
//
// # Error Handling
//
// The package provides typed errors for precise error handling:
//
//	if err := core.Start(config); err != nil {
//	    var portErr *xray.PortError
//	    if errors.As(err, &portErr) {
//	        log.Printf("Port %d is already in use", portErr.Port)
//	        return
//	    }
//
//	    var cfgErr *xray.ConfigError
//	    if errors.As(err, &cfgErr) {
//	        log.Printf("Config error: %s", cfgErr.Reason)
//	        return
//	    }
//
//	    log.Fatal(err)
//	}
//
// # Thread Safety
//
// All public methods of [Core] are safe for concurrent use.
// State reads use RLock for better performance under high concurrency.
package xray
