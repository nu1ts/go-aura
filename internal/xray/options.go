package xray

import "time"

// Option configures a [Core] instance.
// Options are passed to [New] to customize behavior.
//
// Example defining a custom option:
//
//	func WithMetrics(collector MetricsCollector) Option {
//	    return func(c *Core) {
//	        c.metrics = collector
//	    }
//	}
type Option func(*Core)

// WithLogger sets a custom logger implementation.
// If not set, logs are written to stdout.
//
// Example:
//
//	core := xray.New(xray.WithLogger(myLogger))
func WithLogger(l Logger) Option {
	return func(c *Core) {
		if l != nil {
			c.logger = l
		}
	}
}

// WithStartTimeout sets the maximum time to wait for xray-core to start.
// If the timeout is exceeded, [Core.Start] returns [ErrStartTimeout].
//
// Default: 10 seconds.
//
// Example:
//
//	core := xray.New(xray.WithStartTimeout(30 * time.Second))
func WithStartTimeout(d time.Duration) Option {
	return func(c *Core) {
		if d > 0 {
			c.startTimeout = d
		}
	}
}

// WithShutdownTimeout sets the maximum time to wait for graceful shutdown.
// If the timeout is exceeded, the core is forcefully stopped.
//
// Default: 5 seconds.
//
// Example:
//
//	core := xray.New(xray.WithShutdownTimeout(3 * time.Second))
func WithShutdownTimeout(d time.Duration) Option {
	return func(c *Core) {
		if d > 0 {
			c.shutdownTimeout = d
		}
	}
}
