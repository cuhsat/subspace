package sys

import (
	"net"
)

// MaxBuffer is the maximum allowed buffer size
// for sending or receiving UDP packages over IP
// and therefor the maximum size of a signal.
const MaxBuffer = 1024

const (
	Port1 = ":8211" // incoming signal port address.
	Port2 = ":8212" // outgoing signal port address.
)

// NewBuffer returns a signal buffer ready to use.
func NewBuffer() []byte {
	return make([]byte, MaxBuffer)
}

// Address returns the first active MAC address.
//
// Any calling program will terminate immediately if an error occurs.
func Address() []byte {
	li, err := net.Interfaces()
	if err != nil {
		Fatal(err)
	}

	for _, i := range li {
		if (i.Flags & net.FlagUp) != 0 {
			return i.HardwareAddr
		}
	}

	Fatal("no address")

	return nil
}

// Dial opens an UDP connection on the given address,
// sets the internal buffer size for read and write buffers
// and returns an UDP pseudo connection.
//
// Any calling program will terminate immediately if an error occurs.
func Dial(addr string) (u *net.UDPConn) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		Fatal(err)
	}

	u, err = net.DialUDP("udp", nil, a)
	if err != nil {
		Fatal(err)
	}

	u.SetReadBuffer(MaxBuffer)
	u.SetWriteBuffer(MaxBuffer)

	return
}

// Listen opens an UDP listener on the given address,
// sets the internal buffer size for read and write buffers
// and returns an UDP pseudo connection.
//
// Any calling program will terminate immediately if an error occurs.
func Listen(addr string) (u *net.UDPConn) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		Fatal(err)
	}

	u, err = net.ListenUDP("udp", a)
	if err != nil {
		Fatal(err)
	}

	u.SetReadBuffer(MaxBuffer)
	u.SetWriteBuffer(MaxBuffer)

	return
}
