// Ss is a stream cli for subspace server communication.
//
// Outgoing signals will be processed from the standard input.
// Incoming signals will be printed to the standard output,
// followed by a line break after each signal.
// The size of a signal must be between 1 and 1024 bytes.
//
// Usage:
//
//	stdin | ss [host] > stdout
//
// The arguments are:
//
//	host
//		Subspace server host name.
//		Defaults to localhost.
package main

import (
	"fmt"
	"os"

	"github.com/cuhsat/subspace/internal/app/ss"
	"github.com/cuhsat/subspace/internal/pkg/sys"
)

// The main function will open a channel to subspace host
// and will either send or scan signals, dependent on
// there is data to be read from the standard input.
func main() {
	host := "localhost"

	if len(os.Args) > 1 {
		host = os.Args[1]
	}

	c := ss.NewChannel(host)

	b := sys.Stdin()

	if len(b) > sys.MaxBuffer {
		sys.Fatal("buffer overflow")
	} else if len(b) > 0 {
		c.Send(b)
	} else {
		ch := make(chan []byte)

		go c.Scan(ch, sys.Address())

		for v := range ch {
			fmt.Println(string(v))
		}
	}
}
