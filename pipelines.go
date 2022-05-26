package pipelines

import (
	"context"
	"net"
)

type Pipeliner[T any] interface {
	// PipelineChan needs to return a R/W chan that is for the output channel, This is for chaining the pipeline
	PipelineChan() chan T

	// Close is used to clear up any resources made by the component
	Close()
}

type RetryPacket struct {
	// Addr holds a UDP address (with port) for the packet
	// will be ignored if UDP is created in CLIENT mode
	Addr *net.UDPAddr
	// Data contains the data
	Data []byte
	sn   uint64
	ctx  context.Context
	can  context.CancelFunc
}

func (p *RetryPacket) Key() uint64 {
	return p.sn
}

func (p *RetryPacket) Context() context.Context {
	return p.ctx
}

func New(s uint64) *RetryPacket {
	c, cancel := context.WithCancel(context.Background())
	p := &RetryPacket{sn: s, ctx: c, can: cancel}

	return p
}
