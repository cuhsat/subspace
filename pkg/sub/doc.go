// Package sub implements the subspace data structure.
//
// # Subspace Theory
//
// A subspace is data structure of data packets called signals. These signals are strictly ordered by their time of
// arrival in an almost linear fashion (more on that later). All signals will only be held in memory for a defined
// amount of time. After this retention time, the signals will be invalidated called dropped, to be later collected
// by the default garbage collector of Go. All operations on a subspace are safe for concurrent use.
//
// Signals can be retrieved in the present order via a process called scanning. Scanned signals will be written to a
// given Go channel and this channel will be closed afterwards. A state id can be given, to save the state of the last
// Scan for continuing later on. If no state is given, all existing signals will be scanned every time the method is
// called. As this decreases the performance of a subspace significantly, it should be avoided and a state id should
// be used at all times. If a state points to a signal out of retention time, it will be removed automatically with
// the next call of the Drop method.
//
// # Altering Operations
//
// Every time a data structure altering operation (as Send or Drop) is called, the internal operations counter will
// be increased by one. This counter will never decrease. All signal specific methods will return the current count
// as a timestamp of the subspaces current structure.
//
//	s.Send([]byte("foo"))
//	// Will return 1
//
//	s.Scan(make(chan []byte, 1), nil)
//	// Will return 1
//
//	s.Drop(0)
//	// Will return 2
//
// # Internal Clock
//
// A subspace operates its own internal clock with an accuracy of a microsecond. Signals that concurrently arrive at
// the same internal time, will be ordered in the unpredictable way goroutines are processed. The method NewSpace will
// not return until the clocks first tick has occurred. So the minimum duration time of this method is one tenth of a
// millisecond.
//
// # Internal Statistics
//
// Statistic about the subspaces current state can be retrieved via the structures public fields. These fields are not
// safe for concurrent use, not even reading. Please consider to use an atomic operation like LoadUint64 to retrieve a
// value:
//
//	sc := atomic.LoadUint64(&s.StatCount)
//	sa := atomic.LoadUint64(&s.StatAlloc)
//
// There are two public available statistics of a subspace:
//
//  1. StatCount: Number of currently stored signals.
//  2. StatAlloc: Number of currently allocated memory (in bytes).
//
// # No Persistence
//
// As a subspace is a memory only data structure, all signals will not be persisted. If this is required, then the task
// is in the responsibility of the data structures user.
//
// # State Management
//
// A state is a named marker, at which a scan will continue when provided with this name. A state can simply be created
// by using it the first time for a scan. States are not permanent and will be deleted by the Drop method, if they point
// to an already dropped signal. A state can be forked by prefixing its name with an exclamation mark (!). The forked
// state will duplicate its origins marker and will be saved under the forks name (beginning with the exclamation mark).
// A forked state can also be forked. Its name will begin with two exclamation marks (!!). Forking a state is a really
// powerful concept, as it allows a subspace state to be used without altering it.
//
// It is possible to fast forward a state, by simply scanning and ignoring any found signals. It is not possible to
// rewind a state to the first signal of a subspace. If you have to scan signals twice, you should consider forking
// the state beforehand, using a different state name, or using no state (nil) at all.
//
// # Performance Optimizations
//
// To increase the maximum possible performance of a subspace, various aids have been implemented:
//
//  1. A single root signal exists right after its creation. This root signal has an unlimited retention time and will
//     be ignored by any calls to Scan or Drop. It acts as an anchor for all future incoming signals, to be able to
//     fast append them without a performance drop for the fist signal. The last signal in a subspace points to the
//     root signal for faster iteration. Thereby, a subspace is internally a looped directional graph.
//  2. A concurrent internal clock is used to prevent calls to the operating systems time functions directly. Instead,
//     all calls for the current time will use a cached variable with an atomic operation.
//  3. A sync pool for signal structures is used to avoid memory allocations while receiving signals. All used signal
//     structures will return to the structure pool, after they have been dropped. The pool will come into full effect
//     for all received signals, right after a call to the Drop method and before the next garbage collector run.
//
// # License
//
// Usage of this source code is governed by the MIT License. Please see the LICENSE file for further information.
package sub
