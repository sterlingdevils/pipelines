// Package udp_test will test the public API of udp
package pipelines_test

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sterlingdevils/pipelines"
)

const (
	TESTPORT = 9092
)

// Example of how to create,receive and send packets
//
// This will create a UDP component and then send a packet,
// receive the udp, then display it, and check the display is
// correct.
func ExampleUDPPipe_NewWithParams() {
	in := make(chan pipelines.Packetable, 1)

	// Must pass in the input channel as we dont assume we own it
	udpcomp, err := pipelines.UDPPipe{}.NewWithParams(in, ":9092", pipelines.SERVER, 1)
	if err != nil {
		log.Fatalln("error creating UDP")
	}

	// Wait for 1 second, then send a packet to our self, and display it, exit
loopexit:
	for {
		select {
		case <-time.After(time.Second * 1):
			in <- &pipelines.Packet{Addr: net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: TESTPORT}, DataSlice: []byte("Hello from Us.")}
		case p := <-udpcomp.OutChan():
			fmt.Printf("%v: %v\n", p.Address(), p.Data())
			break loopexit
		}
	}

	// Test that when we close we will release the waitgroup
	udpcomp.Close()

	// Close the input channel so we stop reading
	close(in)

	// Output: {127.0.0.1 9092 }: [72 101 108 108 111 32 102 114 111 109 32 85 115 46]
}

func ExampleUDPPipe_New() {
	udpcomp, err := pipelines.UDPPipe{}.New(TESTPORT)
	if err != nil {
		fmt.Printf("failed to create udp component")
	}

	// Send a Packet
	udpcomp.InChan() <- &pipelines.Packet{Addr: net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: TESTPORT}, DataSlice: []byte("Hello from Us.")}

	// Receive a Packet and Display it
	p := <-udpcomp.OutChan()
	fmt.Printf("%v: %v\n", p.Address(), p.Data())

	// Output: {127.0.0.1 9092 }: [72 101 108 108 111 32 102 114 111 109 32 85 115 46]
}
