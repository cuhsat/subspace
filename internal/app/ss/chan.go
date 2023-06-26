// Package ss provides a channel for subspace communications.
package ss

import (
	"net"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cuhsat/subspace/internal/pkg/sys"
)

// A channel is a bi-directional communication provider for a subspace.
type Channel struct {
	Rx uint64       // received bytes.
	Tx uint64       // transmitted bytes.
	ru *net.UDPConn // receiving connection.
	tu *net.UDPConn // transmitting connection.
}

// NewChannel returns a new channel for communicating with a subspace.
// The channel opens two UDP pseudo connections for sending and receiving
// signals as bytes arrays. These connections will be closed automatically
// when the channel is being freed by the garbage collector.
//
// Any calling program will terminate immediately if an error occurs.
func NewChannel(host string) (c *Channel) {
	c = &Channel{
		ru: sys.Dial(host + sys.Port2),
		tu: sys.Dial(host + sys.Port1),
	}

	// automatic close open connections after use
	runtime.SetFinalizer(c, func(c *Channel) {
		c.ru.Close()
		c.tu.Close()
	})

	return
}

// Send the given signal to the subspace via an UDP pseudo connection.
//
// Send will count all transmitted bytes.
func (c *Channel) Send(b []byte) {
	n, err := c.tu.Write(b)

	if err != nil {
		sys.Fatal(err)
	}

	atomic.AddUint64(&c.Tx, uint64(n))
}

// Scan all new signals in a subspace via an UDP pseudo connection.
// If no further signals are received and the deadline of one second is reached,
// we consider the scan finished. So a call has a minimum duration of one second.
//
// Scan will count all received and transmitted bytes.
func (c *Channel) Scan(ch chan<- []byte, state []byte) {
	n, err := c.ru.Write(state)

	if err != nil {
		sys.Fatal(err)
	}

	atomic.AddUint64(&c.Tx, uint64(n))

	for {
		b := sys.NewBuffer()

		c.ru.SetDeadline(time.Now().Add(time.Second))

		n, err := c.ru.Read(b)

		atomic.AddUint64(&c.Rx, uint64(n))

		if os.IsTimeout(err) {
			break // end of scan
		} else if err != nil {
			sys.Fatal(err)
		} else {
			ch <- b[:n]
		}
	}

	close(ch)
}
