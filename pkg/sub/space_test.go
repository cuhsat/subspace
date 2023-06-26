package sub

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cuhsat/subspace/internal/pkg/sys"
)

const (
	_now = -1 // immediately
)

var (
	_foo = []byte("foo")
	_bar = []byte("bar")
)

var _tests = []int{
	1e0, 1e1, 1e2, 1e3,
}

var _s atomic.Pointer[Space]

func Example() {
	s := NewSpace()

	s.Send([]byte("hello"))
	s.Send([]byte("world"))

	ch := make(chan []byte)

	go s.Scan(ch, nil)

	for x := range ch {
		fmt.Println(string(x))
	}

	s.Drop(0)

	// Output:
	// hello
	// world
}

func TestMain(m *testing.M) {
	_cleanup()

	os.Exit(m.Run())
}

func TestNewSpace(t *testing.T) {
	t.Run("NewSpace should return a new space", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := NewSpace()

		if s == nil {
			t.Fatal("Space is nil")
		}

		if s.StatCount != 0 {
			t.Fatal("Count is not zero")
		}

		if s.StatAlloc != 0 {
			t.Fatal("Memory is not zero")
		}
	})
}

func TestSend(t *testing.T) {
	t.Run("Send should increase offset", func(t *testing.T) {
		t.Cleanup(_cleanup)

		o1 := atomic.LoadUint64(&_s.Load().ops)
		o2 := _s.Load().Send([]byte{})

		if o1 == o2 {
			t.Fatal("Offset was not increased")
		}
	})

	t.Run("Send should create a signal", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		s := _s.Load()

		if s.head == s.root {
			t.Fatal("Head points still to root")
		}

		if s.head.time == 0 {
			t.Fatal("Signal time is not set")
		}

		if s.head.data[0] != 1 {
			t.Fatal("Signal data is not correct")
		}

		if s.head.next != s.root {
			t.Fatal("Signal next points not to root")
		}
	})

	t.Run("Send should link in the correct order", func(t *testing.T) {
		t.Cleanup(_cleanup)

		for i := 1; i <= 3; i++ {
			_send(byte(i))
		}

		x1 := _s.Load().root.next
		x2 := x1.next
		x3 := x2.next

		if x1.data[0] != 1 {
			t.Fatal("Signals are not in the correct order")
		}

		if x2.data[0] != 2 {
			t.Fatal("Signals are not in the correct order")
		}

		if x3.data[0] != 3 {
			t.Fatal("Signals are not in the correct order")
		}
	})

	t.Run("Send should update stats", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()

		for i := 1; i <= 3; i++ {
			_send(byte(i))
		}

		_scan(_foo)
		_scan(_bar)

		if s.StatCount != 3 {
			t.Fatal("Signal count is not correct")
		}

		if s.StatAlloc != 3 {
			t.Fatal("Memory is not correct")
		}
	})
}

func TestScan(t *testing.T) {
	t.Run("Scan should not change offset", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		ch := make(chan []byte, 1)

		o1 := atomic.LoadUint64(&_s.Load().ops)
		o2 := _s.Load().Scan(ch, _foo)

		if o1 != o2 {
			t.Fatal("Offset was changed")
		}
	})

	t.Run("Scan should return nothing for root", func(t *testing.T) {
		t.Cleanup(_cleanup)

		ch := make(chan []byte, 1)

		_s.Load().Scan(ch, _foo)

		if len(ch) > 0 {
			t.Fatal("Channel was not empty")
		}
	})

	t.Run("Scan should return nothing when empty", func(t *testing.T) {
		t.Cleanup(_cleanup)

		v := _scan(_foo)

		if len(v) > 0 {
			t.Fatal("Data is not empty")
		}
	})

	t.Run("Scan should return nothing if fully scanned", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		v1 := _scan(_foo)
		v2 := _scan(_foo)

		if len(v1) < 1 {
			t.Fatal("Data is not correct")
		}

		if len(v2) > 0 {
			t.Fatal("Data is not empty")
		}
	})

	t.Run("Scan should return the correct order", func(t *testing.T) {
		t.Cleanup(_cleanup)

		for n := 1; n <= 3; n++ {
			_send(byte(n))
		}

		v := _scan(_foo)

		for n := 0; n < 3; n++ {
			if v[n] != byte(n+1) {
				t.Fatal("Data is not correct")
			}
		}
	})

	t.Run("Scan should continue where it ended", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)
		_send(2)
		_send(3)

		_scan(_foo)

		_send(4)

		v := _scan(_foo)

		if v[0] < 4 {
			t.Fatal("State has not continued")
		}
	})

	t.Run("Scan should return everything for state nil", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		ch1 := make(chan []byte, 1)

		_s.Load().Scan(ch1, nil)

		ch2 := make(chan []byte, 1)

		_s.Load().Scan(ch2, nil)

		if (<-ch2)[0] != 1 {
			t.Fatal("Data is not correct")
		}
	})

	t.Run("Scan should move the state", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()

		if s.states.m[string(_foo)] != nil {
			t.Fatal("State points not to nil")
		}

		_send(1)
		_scan(_foo)

		if s.states.m[string(_foo)] != s.head {
			t.Fatal("State points not to head")
		}

		_send(2)
		_scan(_foo)

		if s.states.m[string(_foo)] != s.head {
			t.Fatal("State points not to head")
		}
	})

	t.Run("Scan should fork the state", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s1 := []byte("test")
		s2 := []byte("!test")

		_send(1)

		_scan(s1)
		_scan(s2)

		_send(2)

		_scan(s1)

		s := _s.Load()

		if s.states.m[string(s1)].next != s.root {
			t.Fatal("State next points not to root")
		}

		if s.states.m[string(s2)].next != s.head {
			t.Fatal("State next points not to head")
		}
	})

	t.Run("Scan should update stats", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()

		_send(1)

		_scan(_foo)
		_scan(_bar)

		if s.StatCount != 1 {
			t.Fatal("Count is not correct")
		}

		if s.StatAlloc != 1 {
			t.Fatal("Memory is not correct")
		}
	})

	t.Run("Scan should reset the state if dropped", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		_scan(_foo)
		_drop()
		_scan(_foo)

		_send(2)

		s := _s.Load()

		s.head.time = Infinite

		x := s.states.m[string(_foo)]

		if x != s.root {
			t.Fatal("State was not reset")
		}

		if x.next != s.head {
			t.Fatal("State next is not on head")
		}
	})
}

func TestDrop(t *testing.T) {
	t.Run("Drop should change offset", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		o1 := atomic.LoadUint64(&_s.Load().ops)
		o2 := _drop()

		if o1 == o2 {
			t.Fatal("Offset was not changed")
		}
	})

	t.Run("Drop should not remove root signal", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_drop()

		if _s.Load().root == nil {
			t.Fatal("Root is nil")
		}
	})

	t.Run("Drop should not move forward root", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		x := _s.Load().root

		_drop()

		if x != _s.Load().root {
			t.Fatal("Root moved forward")
		}
	})

	t.Run("Drop should stop for newer signals", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()

		_send(1)
		s.head.time = 0

		_send(2)
		s.head.time = Infinite

		_drop()

		if s.root.next == s.root {
			t.Fatal("All signals were dropped")
		}

		if s.root.next.data[0] < byte(2) {
			t.Fatal("No signals were dropped")
		}
	})

	t.Run("Drop should reset head", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		time.Sleep(time.Second)

		_drop()

		s := _s.Load()

		if s.head != s.root {
			t.Fatal("Head is not on root")
		}

		if s.root.time != Infinite {
			t.Fatal("Root is not root")
		}

		if s.root.next != s.root {
			t.Fatal("Root next is not root")
		}
	})

	t.Run("Drop should delete invalid states", func(t *testing.T) {
		t.Cleanup(_cleanup)

		_send(1)

		time.Sleep(time.Second)

		_scan(_foo)
		_drop()

		_, ok := _s.Load().states.m[string(_foo)]

		if ok {
			t.Fatal("State was not deleted")
		}
	})

	t.Run("Drop should update stats", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()

		for i := 1; i <= 3; i++ {
			_send(byte(i))
		}

		_scan(_foo)
		_scan(_bar)

		_drop()

		if s.StatCount != 0 {
			t.Fatal("Count is not zero")
		}

		if s.StatAlloc != 0 {
			t.Fatal("Memory is not zero")
		}
	})

	t.Run("Drop should lock", func(t *testing.T) {
		t.Cleanup(_cleanup)

		for i := 0; i < 1000000; i++ {
			_send(byte(i))
		}

		time.Sleep(time.Second)

		var wg sync.WaitGroup

		wg.Add(3)

		go func() { _drop(); wg.Done() }()
		go func() { _drop(); wg.Done() }()
		go func() { _drop(); wg.Done() }()

		wg.Wait()

		s := _s.Load()

		if s.root != s.head {
			t.Fatal("Root is not on head")
		}

		if s.root.time != Infinite {
			t.Fatal("Root is not root")
		}

		if s.root.next != s.root {
			t.Fatal("Root next is not root")
		}
	})
}

func BenchmarkNewSpace(b *testing.B) {
	b.Run("Benchmark NewSpace", func(b *testing.B) {
		b.Cleanup(_cleanup)

		for n := 0; n < b.N; n++ {
			NewSpace()
		}
	})
}

func BenchmarkSend(b *testing.B) {
	for _, m := range _tests {
		b.Run(fmt.Sprintf("Benchmark Send %d", m), func(b *testing.B) {
			b.Cleanup(_cleanup)

			s := _s.Load()
			d := sys.NewBuffer()

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				for i := 0; i < m; i++ {
					s.Send(d)
				}
			}
		})
	}
}

func BenchmarkScan(b *testing.B) {
	for _, m := range _tests {
		b.Run(fmt.Sprintf("Benchmark Scan %d", m), func(b *testing.B) {
			b.Cleanup(_cleanup)

			s := _s.Load()
			d := sys.NewBuffer()

			for i := 0; i < m; i++ {
				s.Send(d)
			}

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				ch := make(chan []byte, m)
				st := []byte(strconv.Itoa(n))

				s.Scan(ch, st)
			}
		})
	}
}

func BenchmarkDrop(b *testing.B) {
	for _, m := range _tests {
		b.Run(fmt.Sprintf("Benchmark Drop %d", m), func(b *testing.B) {
			b.Cleanup(_cleanup)

			s := _s.Load()
			d := sys.NewBuffer()

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				for i := 0; i < m; i++ {
					s.Send(d)
				}

				s.Drop(_now)
			}
		})
	}
}

func _send(b byte) uint64 {
	return _s.Load().Send([]byte{b})
}

func _scan(b []byte) []byte {
	bs := make([]byte, 0)
	ch := make(chan []byte)

	go _s.Load().Scan(ch, b)

	for v := range ch {
		bs = append(bs, v[0])
	}

	return bs
}

func _drop() uint64 {
	return _s.Load().Drop(_now)
}

func _cleanup() {
	_s.Swap(NewSpace())
}
