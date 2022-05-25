package logpipe

import (
	"context"
	"log"

	"github.com/sterlingdevils/pipelines/pkg/pipeline"
)

const (
	CHANSIZE = 0
)

type LogPipe[T any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan T
	outchan chan T

	pl pipeline.Pipelineable[T]
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b *LogPipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *LogPipe[_]) Close() {
	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()
}

// mainloop, read from in channel and write to out channel safely, log the item
// exit when our context is closed
func (b *LogPipe[_]) mainloop() {
	defer close(b.outchan)

	for {
		select {
		case t := <-b.inchan:
			log.Println(t)
			select {
			case b.outchan <- t:
			case <-b.ctx.Done():
				return
			}
		case <-b.ctx.Done():
			return
		}
	}
}

func NewWithChannel[T any](in chan T) *LogPipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := LogPipe[T]{ctx: con, can: cancel, inchan: in, outchan: make(chan T, CHANSIZE)}

	go r.mainloop()

	return &r
}

func NewWithPipeline[T any](p pipeline.Pipelineable[T]) *LogPipe[T] {
	r := NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

func New[T any]() *LogPipe[T] {
	return NewWithChannel(make(chan T, CHANSIZE))
}
