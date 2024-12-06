package subspace

import (
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cuhsat/subspace/internal/pkg/sys"
	"github.com/cuhsat/subspace/pkg/sub"
)

var (
	_foo = []byte("foo")
	_bar = []byte("bar")
)

var _s atomic.Pointer[sub.Space]

func TestMain(m *testing.M) {
	_cleanup()

	os.Exit(m.Run())
}

func TestRelay(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip() // Faulty CI
	}

	t.Run("Relay should relay a signal to a relay", func(t *testing.T) {
		t.Cleanup(_cleanup)

		go Relay([]string{"localhost"})

		s := _s.Load()
		u := sys.Listen("localhost" + sys.Port1)

		defer u.Close()

		go Send(u, s)

		_sendOnce()

		time.Sleep(time.Millisecond)

		if atomic.LoadUint64(&Fx) == 0 {
			t.Fatal("Signal was not relayed")
		}
	})
}

func TestSend(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip() // Faulty CI
	}

	t.Run("Send should send a signal to the subspace", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()
		u := sys.Listen("localhost" + sys.Port1)

		defer u.Close()

		go Send(u, s)

		_sendOnce()

		time.Sleep(time.Millisecond)

		if atomic.LoadUint64(&s.StatCount) == 0 {
			t.Fatal("Signal was not send")
		}
	})
}

func TestScan(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip() // Faulty CI
	}

	t.Run("Scan should scan a signal from the subspace", func(t *testing.T) {
		t.Cleanup(_cleanup)

		s := _s.Load()
		u := sys.Listen("localhost" + sys.Port2)

		defer u.Close()

		go Scan(u, s)

		s.Send(_foo)

		b := _scanOnce()

		time.Sleep(time.Millisecond)

		if len(b) == 0 {
			t.Fatal("Signal was not scanned")
		}
	})
}

func BenchmarkRelay(b *testing.B) {
	b.Run("Benchmark Relay", func(b *testing.B) {
		b.Cleanup(_cleanup)

		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			Relay([]string{"localhost"})
		}

		b.StopTimer()
	})
}

func BenchmarkSend(b *testing.B) {
	b.Run("Benchmark Send", func(b *testing.B) {
		b.Cleanup(_cleanup)

		s := _s.Load()
		u := sys.Listen("localhost" + sys.Port1)

		defer u.Close()

		loop := true

		go _sendLoop(&loop)

		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			Send(u, s)
		}

		b.StopTimer()

		loop = false
	})
}

func BenchmarkScan(b *testing.B) {
	b.Run("Benchmark Scan", func(b *testing.B) {
		b.Cleanup(_cleanup)

		s := _s.Load()
		u := sys.Listen("localhost" + sys.Port2)

		defer u.Close()

		loop := true

		go _scanLoop(&loop)

		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			Scan(u, s)
		}

		b.StopTimer()

		loop = false
	})
}

func _sendOnce() {
	u := sys.Dial("localhost" + sys.Port1)

	defer u.Close()

	if _, err := u.Write(_foo); err != nil {
		panic(err)
	}
}

func _sendLoop(loop *bool) {
	u := sys.Dial("localhost" + sys.Port1)

	for *loop {
		u.Write(_foo)
	}

	u.Close()
}

func _scanOnce() (b []byte) {
	u := sys.Dial("localhost" + sys.Port2)

	defer u.Close()

	b = sys.NewBuffer()

	if _, err := u.Write(_bar); err != nil {
		panic(err)
	}

	if _, err := u.Read(b); err != nil {
		panic(err)
	}

	return
}

func _scanLoop(loop *bool) {
	u := sys.Dial("localhost" + sys.Port2)

	for *loop {
		u.Write(_bar)
	}

	u.Close()
}

func _cleanup() {
	_s.Swap(sub.NewSpace())
}
