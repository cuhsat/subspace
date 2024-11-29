// Subspace is a memory only subspace server.
//
// The server will run until an exit signal either of SIGINT or SIGTERM is triggered.
// Its stats will be logged to the file system under /tmp/subspace in JSON format.
//
// Usage:
//
//	subspace [retention]
//
// The arguments are:
//
//	retention
//		Signal retention time in seconds.
//		Defaults to 3600 seconds (1 hour).
//
// For communication, two UDP network ports will be opened listening:
//   - 8211 for incoming signals.
//   - 8212 for outgoing signals.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cuhsat/subspace/internal/app/subspace"
	"github.com/cuhsat/subspace/internal/pkg/sys"
	"github.com/cuhsat/subspace/pkg/sub"
)

// The main function will create a new subspace and binds it to two relay routines,
// waiting for incoming pseudo-connections to send or scan signals.
// It will run its own signal garbage collection periodic in the background.
func main() {
	rt := int(time.Hour / 1e9)

	if len(os.Args) > 1 {
		rt, _ = strconv.Atoi(os.Args[1])
	}

	s := sub.NewSpace()

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	go bind(s, subspace.Send, sys.Port1)
	go bind(s, subspace.Scan, sys.Port2)

	go gc(s, rt)

	fmt.Printf("⇌ Subspace [%ds]\n", rt)

	<-exit

	fmt.Printf("⇌ Subspace lost\n")
}

// Bind the given network address to a subspace relay routine.
// The given relay routine will be called until the program exits.
func bind(s *sub.Space, relay subspace.Relay, addr string) {
	u := sys.Listen(addr)

	defer u.Close()

	for {
		relay(u, s)
	}
}

// GC triggers the subspace garbage collection per drop every second
// and logs stats about the space and its traffic as JSON
// to the stats output, overwriting it each time.
func gc(s *sub.Space, rt int) {
	for range time.Tick(time.Second) {
		if rt > 0 {
			s.Drop(int64(rt) * 1e3)
		}

		j, err := json.Marshal(struct {
			Num, Mem, Rx, Tx uint64
		}{
			atomic.LoadUint64(&s.StatCount),
			atomic.LoadUint64(&s.StatAlloc),
			atomic.LoadUint64(&subspace.Rx),
			atomic.LoadUint64(&subspace.Tx),
		})

		if err == nil {
			sys.Stats.Truncate(0)
			sys.Stats.Seek(0, 0)

			fmt.Fprintf(sys.Stats, "%s\n", j)
		}
	}
}
