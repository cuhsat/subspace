package sub

import (
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

// Infinite retention time
const Infinite = math.MaxInt64

// NewSpace returns a new Space struct with its fields initialized.
//
// Each space contains its own pool for signal structures.
//
// NewSpace will only return if the spaces internal clock is running.
// So a call has minimum duration time of one tenth of a millisecond.
func NewSpace() (s *Space) {
	s = &Space{
		states: &states{m: make(map[string]*signal)},
		root:   &signal{Infinite, nil, nil},
		pool: sync.Pool{
			New: func() any {
				return &signal{next: s.root}
			},
		},
	}

	// initialize internal clock
	go s.clock()

	// form the space to a circle
	s.head, s.root.next = s.root, s.root

	// wait till internal clock is running
	for atomic.LoadInt64(&s.now) == 0 {
		runtime.Gosched()
	}

	return
}

// Send will append the given signal at the end of the space.
//
// While the signal is appended, the space will be locked.
//
// Send will return the current spaces operations count
// as a timestamp of the spaces internal signal state.
func (s *Space) Send(data []byte) uint64 {
	x := s.pool.Get().(*signal)

	x.time, x.data = atomic.LoadInt64(&s.now), data

	// lock for fast append
	s.Lock()
	s.head.next, s.head = x, x
	s.Unlock()

	atomic.AddUint64(&s.StatAlloc, uint64(len(data)))
	atomic.AddUint64(&s.StatCount, 1)

	return atomic.AddUint64(&s.ops, 1)
}

// Scan all signals since the beginning or since the given state.
// The given channel will be closed. If the state does not exists,
// it will be created. If a state begins with an '!', the state
// without the exclamation mark will be forked and saved under the
// new name. Forks are no different from normal states and will be
// removed, if they point to a dropped signal. Forks can also be
// forked again.
//
// This should be run as a goroutine or a big enough channel must
// be provided, since this is a blocking call.
//
// Scan will return the current spaces operations count
// as a timestamp of the spaces internal signal state.
func (s *Space) Scan(ch chan<- []byte, state []byte) uint64 {
	k := state

	// fork state if prefixed
	if len(state) > 0 && state[0] == '!' {
		k = state[1:]
	}

	s.states.RLock()
	x, ok := s.states.m[string(k)]
	s.states.RUnlock()

	s.RLock()

	// skip if already on head
	if x != s.head {

		// no state or state was dropped
		if !ok || x.data == nil {
			x = s.root
		}

		// iterate through all signals until head
		for x = x.next; x != s.root; x = x.next {
			ch <- x.data
		}
	}

	// save state if given
	if state != nil {
		s.states.Lock()
		s.states.m[string(state)] = s.head
		s.states.Unlock()
	}

	s.RUnlock()

	close(ch)

	return atomic.LoadUint64(&s.ops)
}

// Drop invalidates all signals older than the given retention time.
// This is done by setting the signals data pointer to nil,
// so that the garbage collector will remove it afterwards.
//
// Drop will also remove all states that point to an invalid signal.
//
// While the signals are invalidated, the space will be locked.
//
// Drop will return the current spaces operations count
// as a timestamp of the spaces internal signal state.
func (s *Space) Drop(retention int64) uint64 {
	t, o := atomic.LoadInt64(&s.now)-retention, uint64(0)

	s.RLock()
	x := s.root.next
	s.RUnlock()

	s.Lock()

	// invalidate all signals until new enough
	for n := x; t > x.time; x, o = n, 1 {
		atomic.AddUint64(&s.StatAlloc, ^uint64(len(x.data)-1))
		atomic.AddUint64(&s.StatCount, ^uint64(0))

		n, x.data, x.next = x.next, nil, s.root

		s.pool.Put(x)
	}

	// reset head if invalid
	if s.head.data == nil {
		s.head = s.root
	}

	s.root.next = x

	s.Unlock()

	s.states.Lock()

	// drop now invalid states
	for k, v := range s.states.m {
		if v.data == nil {
			delete(s.states.m, k)
		}
	}

	s.states.Unlock()

	return atomic.AddUint64(&s.ops, o)
}
