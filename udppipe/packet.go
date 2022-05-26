package udppipe

import "net"

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
