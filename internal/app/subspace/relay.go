// Package subspace provides relay routines for subspace communications.
package subspace

import (
	"net"
	"sync/atomic"

	"github.com/cuhsat/subspace/internal/pkg/sys"
	"github.com/cuhsat/subspace/pkg/sub"
)

// Relay routines provided for binding
type Relay func(u *net.UDPConn, s *sub.Space)

var (
	Rx uint64 // received bytes.
	Tx uint64 // transmitted bytes.
)

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
