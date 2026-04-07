package xray

import "fmt"

// Logger defines the minimal logging interface.
// This allows integration with any logging library (zerolog, slog, logrus, etc.).
//
// Example implementation for Wails:
//
//	type WailsLogger struct {
//	    ctx context.Context
//	}
//
//	func (l *WailsLogger) Info(msg string, args ...interface{}) {
//	    runtime.LogInfof(l.ctx, msg, args...)
//	}
//
//	func (l *WailsLogger) Error(msg string, args ...interface{}) {
//	    runtime.LogErrorf(l.ctx, msg, args...)
//	}
//
//	func (l *WailsLogger) Debug(msg string, args ...interface{}) {
//	    runtime.LogDebugf(l.ctx, msg, args...)
//	}
//
// Example implementation for zerolog:
//
//	type ZerologAdapter struct {
//	    log zerolog.Logger
//	}
//
//	func (l *ZerologAdapter) Info(msg string, args ...interface{}) {
//	    l.log.Info().Msgf(msg, args...)
//	}
//
//	func (l *ZerologAdapter) Error(msg string, args ...interface{}) {
//	    l.log.Error().Msgf(msg, args...)
//	}
//
//	func (l *ZerologAdapter) Debug(msg string, args ...interface{}) {
//	    l.log.Debug().Msgf(msg, args...)
//	}
type Logger interface {
	// Info logs informational messages.
	Info(msg string, args ...interface{})

	// Error logs error messages.
	Error(msg string, args ...interface{})

	// Debug logs debug messages.
	Debug(msg string, args ...interface{})
}

// defaultLogger is a fallback implementation that writes to stdout.
// Used when no custom logger is provided via [WithLogger].
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO]  "+msg+"\n", args...)
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

// NopLogger is a logger that discards all messages.
// Useful for testing or when logging is not needed.
//
// Example:
//
//	core := xray.New(xray.WithLogger(xray.NopLogger{}))
type NopLogger struct{}

func (NopLogger) Info(_ string, _ ...interface{})  {}
func (NopLogger) Error(_ string, _ ...interface{}) {}
func (NopLogger) Debug(_ string, _ ...interface{}) {}
