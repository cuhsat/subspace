package ss

import (
	"bytes"
	"os"
	"testing"

	"github.com/cuhsat/subspace/internal/pkg/sys"
)

var (
	_host = "localhost"
	_ping = []byte("ping")
)

func TestMain(m *testing.M) {
	go _echo(sys.Port1)
	go _echo(sys.Port2)

	os.Exit(m.Run())
}

func TestNewChannel(t *testing.T) {
	t.Run("NewChannel should return a new channel", func(t *testing.T) {
		c := NewChannel(_host)

		if c == nil {
			t.Fatal("Channel is nil")
		}
	})
}

func TestSend(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip() // Faulty CI
	}

	t.Run("Send should send a signal", func(t *testing.T) {
		c := NewChannel(_host)

		c.Send(_ping)

		if c.Tx != uint64(len(_ping)) {
			t.Fatal("Tx is wrong")
		}

		if c.Rx != 0 {
			t.Fatal("Rx is wrong")
		}
	})
}

func TestScan(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip() // Faulty CI
	}

	t.Run("Scan should scan a signal", func(t *testing.T) {
		c := NewChannel(_host)

		ch := make(chan []byte, 1)

		c.Scan(ch, _ping)

		if c.Tx != uint64(len(_ping)) {
			t.Fatal("Tx is wrong")
		}

		if c.Rx != uint64(len(_ping)) {
			t.Fatal("Rx is wrong")
		}

		if !bytes.Equal(<-ch, _ping) {
			t.Fatal("Data is not correct")
		}
	})
}

func BenchmarkNewChannel(b *testing.B) {
	b.Run("Benchmark NewChannel", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			NewChannel(_host)
		}
	})
}

func BenchmarkSend(b *testing.B) {
	b.Run("Benchmark Send", func(b *testing.B) {
		c := NewChannel(_host)

		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			c.Send(_ping)
		}
	})
}

func BenchmarkScan(b *testing.B) {
	b.Run("Benchmark Scan", func(b *testing.B) {
		c := NewChannel(_host)

		ch := make(chan []byte)

		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			c.Scan(ch, nil)
		}
	})
}

func _echo(port string) {
	b := sys.NewBuffer()
	u := sys.Listen(_host + port)

	defer u.Close()

	n, addr, err := u.ReadFromUDP(b)

	if err != nil {
		panic(err)
	}

	_, err = u.WriteToUDP(b[:n], addr)

	if err != nil {
		panic(err)
	}
}
