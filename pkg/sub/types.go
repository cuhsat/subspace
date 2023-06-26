package sub

import (
	"sync"
)

// A space represents a chronological order of signals.
// Root and head fields are not safe for concurrent usage
// and the space must be locked for every access to them.
//
// Public stats are not safe for concurrent usage.
//
// A space should never be reused.
type Space struct {
	sync.RWMutex
	// Current count of signals.
	StatCount uint64
	// Current allocated memory.
	StatAlloc uint64
	// Current space time.
	now int64
	// Every time a space altering operation happens,
	// the ops value will be increased by one.
	// The ops value will never be decreased.
	ops uint64
	// The root is a performance optimization for faster appending new signals.
	// All signals will be chained from it. Because of its infinite signal time,
	// it will never be dropped.
	root *signal
	// The head will always point to the last signal of the chain.
	// It will point to the root signal initially.
	head *signal
	// Pool of cached signal structures.
	pool sync.Pool
	// Storage of scan states.
	states *states
}

// States is a lockable storage for scan states.
// The underlying map is not safe for concurrent usage
// and the state must be locked for every access to it.
//
// A state points to the last signal that was scanned.
//
// State entries are non-persistent and will be deleted
// by the time the signal they are pointing to is dropped.
type states struct {
	sync.RWMutex
	m map[string]*signal // Underlying map.
}

// A signal represents a received data package.
// Signals are chained forward-only by their next field
// in strict ascending chronological order of the time field.
//
// A signal with nil as value for data or next has been dropped
// and will be consumed by the garbage collector soon.
type signal struct {
	time int64   // Time of receiving.
	data []byte  // Received data.
	next *signal // Next signal.
}
