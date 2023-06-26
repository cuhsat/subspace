// Proxy is subspace proxy server.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cuhsat/subspace/internal/app/ss"
)

const mime = "application/json"
const host = "localhost"
const port = ":8080"

type signals struct {
	Signals [][]byte
}

func main() {
	c := ss.NewChannel(host)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusMethodNotAllowed

		switch r.Method {
		case http.MethodPost:
			code = send(c, w, r)
		case http.MethodGet:
			code = scan(c, w, r)
		}

		if code != http.StatusOK {
			http.Error(w, "Error", code)
		}
	})

	fmt.Printf("⇌ Subspace Proxy%s\n", port)

	http.ListenAndServe(port, mux)
}

func send(c *ss.Channel, w http.ResponseWriter, r *http.Request) int {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		return http.StatusInternalServerError
	}

	c.Send(b)

	return http.StatusOK
}

func scan(c *ss.Channel, w http.ResponseWriter, r *http.Request) int {
	ch := make(chan []byte)

	var state []byte
	if len(r.URL.Path) > 0 {
		state = []byte(r.URL.Path)
	}

	go c.Scan(ch, state)

	var s signals
	for x := range ch {
		s.Signals = append(s.Signals, x)
	}

	w.Header().Set("Content-Type", mime)

	json.NewEncoder(w).Encode(s)

	return http.StatusOK
}
