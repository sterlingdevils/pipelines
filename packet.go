package pipelines

import (
	"net"

	"github.com/sterlingdevils/gobase/serialnum"
)

type Packetable interface {
	Address() net.UDPAddr
	Data() []byte
	// DataLength() int
}

// Packet holds a UDP address and Data from the UDP
// The channel types for the input and output channel are of this type
type Packet struct {
	// Addr holds a UDP address (with port) for the packet
	// will be ignored if UDP is created in CLIENT mode
	Addr net.UDPAddr
	// Data contains the data
	DataSlice []byte
}

func (p Packet) Address() net.UDPAddr {
	return p.Addr
}

func (p Packet) Data() []byte {
	return p.DataSlice
}

func (p Packet) Size() int {
	return len(p.DataSlice)
}

// KeyablePacket holds a UDP address and Data from the UDP
// The channel types for the input and output channel are of this type
type KeyablePacket struct {
	// Addr holds a UDP address (with port) for the packet
	// will be ignored if UDP is created in CLIENT mode
	Addr net.UDPAddr
	// Data contains the data
	DataSlice []byte
}

func (p KeyablePacket) Key() uint64 {
	s, _ := serialnum.Uint64(p.DataSlice)
	return s
}

func (p KeyablePacket) Address() net.UDPAddr {
	return p.Addr
}

func (p KeyablePacket) Data() []byte {
	return p.DataSlice
}

func (p KeyablePacket) Size() int {
	return len(p.DataSlice)
}
