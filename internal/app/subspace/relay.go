// Package subspace provides relay routines for subspace communications.
package subspace

import (
	"net"
  "runtime"
	"sync/atomic"

	"github.com/cuhsat/subspace/internal/pkg/sys"
	"github.com/cuhsat/subspace/pkg/sub"
)

// Bindable routines definition.
type Bind func(u *net.UDPConn, s *sub.Space)

var (
  // Received bytes.
  Rx uint64
	// Transmitted bytes.
  Tx uint64
  // Forwarded bytes.
  Fx uint64
  // Data channel. 
  dc atomic.Pointer[chan []byte]
)

// A relay is a uni-directional communication relay to another subspace relay.
type relay struct {
	tu *net.UDPConn // transmitting connection.
}

// NewRelay returns a new relay for forwarding signals to a subspace.
// The relay opens a UDP pseudo connection for sending
// signals as bytes arrays. This connection will be closed automatically
// when the relay is being freed by the garbage collector.
//
// Any calling program will terminate immediately if an error occurs.
func NewRelay(host string) (r *relay) {
	r = &relay{
		tu: sys.Dial(host + sys.Port1),
	}

	// automatic close open connections after use
	runtime.SetFinalizer(r, func(r *relay) {
		r.tu.Close()
	})

	return
}

// Relay forwards all received signal data the to given relays.
//
// Relay will count all transmitted bytes.
//
// Any calling program will terminate immediately if an error occurs.
func Relay(hosts []string) {
  rs := make([]*relay, 0)

  for _, host := range hosts {
    rs = append(rs, NewRelay(host))
  }
 
  ch := make(chan []byte)

  dc.Store(&ch)
  
  go func() {
    for b := range *dc.Load() {
      for _, r := range rs {
        n, err := r.tu.Write(b)

	      if err != nil {
		      sys.Fatal(err)
	      }

	      atomic.AddUint64(&Fx, uint64(n))
      }
    } 
  }()
}

// Send receives data from an UDP pseudo connection
// and send this data as a signal to the given subspace.
//
// Send will count all received bytes.
func Send(u *net.UDPConn, s *sub.Space) {
	b := sys.NewBuffer()

	n, _, err := u.ReadFromUDP(b)

	if err == nil {
		go s.Send(b[:n])
	}

	atomic.AddUint64(&Rx, uint64(n))

  if c := dc.Load(); c != nil {
    *c <- b[:n]
  }
}

// Scan receives a state id from an UDP pseudo connection
// and scans the given subspace using the id for new signals.
//
// Scanned signals are send in parallel to the received address.
//
// Scan will count all received and transmitted bytes.
func Scan(u *net.UDPConn, s *sub.Space) {
	var b [sys.MaxBuffer]byte

	n, addr, err := u.ReadFromUDP(b[:])

	atomic.AddUint64(&Rx, uint64(n))

	if err == nil {
		ch := make(chan []byte)

		go s.Scan(ch, b[:n])

		go func() {
			for v := range ch {
				n, _ := u.WriteToUDP(v, addr)

				atomic.AddUint64(&Tx, uint64(n))
			}
		}()
	}
}
