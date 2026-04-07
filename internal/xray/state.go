package xray

// State represents the current state of xray-core as a finite state machine.
//
// State transitions:
//
//	┌─────────┐   Start()    ┌──────────┐   success   ┌─────────┐
//	│ Stopped │─────────────►│ Starting │────────────►│ Running │
//	│         │◄─────────────│          │◄────────────│         │
//	└─────────┘              └──────────┘   failure   └────┬────┘
//	     ▲                                                 │
//	     │                   ┌──────────┐    Stop()        │
//	     └───────────────────│ Stopping │◄─────────────────┘
//	                         └──────────┘
type State int

const (
	// StateStopped indicates the core is stopped and ready to start.
	StateStopped State = iota

	// StateStarting indicates the core is in the process of starting.
	StateStarting

	// StateRunning indicates the core is running and processing traffic.
	StateRunning

	// StateStopping indicates the core is in the process of stopping.
	StateStopping
)

// String implements [fmt.Stringer] for human-readable state output.
//
// Example:
//
//	fmt.Println(xray.StateRunning) // Output: "running"
func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// IsActive reports whether the state represents an active or transitioning core.
// Returns true for StateStarting, StateRunning, and StateStopping.
func (s State) IsActive() bool {
	return s == StateStarting || s == StateRunning || s == StateStopping
}
