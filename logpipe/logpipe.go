package logpipe

import (
	"context"
	"log"
	"reflect"
	"sync"

	"github.com/sterlingdevils/pipelines"
)

const (
	CHANSIZE = 0
)

type LogPipe[T any] struct {
	name string
	ctx  context.Context
	can  context.CancelFunc

	inchan  chan T
	outchan chan T

	pl pipelines.Pipeline[T]
	wg sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b *LogPipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *LogPipe[_]) Close() {
	defer log.Printf("<logpipe %v> finishing Close call\n", b.name)

	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()

	// Wait for us to be done
	b.wg.Wait()
}

// mainloop, read from in channel and write to out channel safely, log the item
// exit when our context is closed
func (b *LogPipe[_]) mainloop() {
	defer b.wg.Done()
	defer close(b.outchan)
	defer log.Printf("<logpipe %v> closing output channel\n", b.name)

	for {
		select {
		case t, ok := <-b.inchan:
			if !ok {
				return
			}
			log.Printf("<logpipe %v> type:%v   value:%v\n", b.name, reflect.TypeOf(t), t)
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

func NewWithChannel[T any](name string, in chan T) *LogPipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := LogPipe[T]{name: name, ctx: con, can: cancel, inchan: in, outchan: make(chan T, CHANSIZE)}
	log.Printf("<logpipe %v> created\n", name)

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func NewWithPipeline[T any](name string, p pipelines.Pipeline[T]) *LogPipe[T] {
	r := NewWithChannel(name, p.PipelineChan())
	r.pl = p
	log.Printf("<logpipe %v> pipeline set\n", name)
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func New[T any](name string) *LogPipe[T] {
	return NewWithChannel(name, make(chan T, CHANSIZE))
}
