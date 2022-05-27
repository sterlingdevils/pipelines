package udppipe

import (
	"net"

	"github.com/sterlingdevils/gobase/pkg/serialnum"
)

// Packet holds a UDP address and Data from the UDP
// The channel types for the input and output channel are of this type
type Packet struct {
	// Addr holds a UDP address (with port) for the packet
	// will be ignored if UDP is created in CLIENT mode
	Addr *net.UDPAddr
	// Data contains the data
	Data []byte
}

func (p *Packet) Size() int {
	return len(p.Data)
}

// Packet holds a UDP address and Data from the UDP
// The channel types for the input and output channel are of this type
type KeyablePacket struct {
	// Addr holds a UDP address (with port) for the packet
	// will be ignored if UDP is created in CLIENT mode
	Addr *net.UDPAddr
	// Data contains the data
	Data []byte
}

func (p *KeyablePacket) Key() uint64 {
	s, _ := serialnum.Uint64(p.Data)
	return s
}

func (p *KeyablePacket) Size() int {
	return len(p.Data)
}
