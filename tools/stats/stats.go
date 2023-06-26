// Stats is a subspace stats server.
package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cuhsat/subspace/internal/pkg/sys"
)

const port = ":8081"

type cache struct {
	time time.Time
	data []byte
}

func main() {
	var cache cache

	l, err := net.Listen("tcp", port)

	if err != nil {
		sys.Fatal(err)
	}

	defer l.Close()

	fmt.Printf("⇌ Subspace Stats%s\n", port)

	for {
		c, err := l.Accept()

		if err != nil {
			continue
		}

		b, err := cache.sync()

		if err == nil {
			c.Write(b)
		}

		c.Close()
	}
}

func (c *cache) sync() (b []byte, err error) {
	fi, err := sys.Stats.Stat()

	if err != nil {
		return
	}

	mt := fi.ModTime()

	if mt.After(c.time) || c.time.IsZero() {
		c.time = mt
		c.data, err = io.ReadAll(sys.Stats)
	}

	b = c.data

	return
}
