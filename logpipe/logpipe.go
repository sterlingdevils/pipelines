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
	ctx context.Context
	can context.CancelFunc

	inchan  chan T
	outchan chan T

	pl pipelines.Pipeliner[T]
	wg sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b *LogPipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *LogPipe[_]) Close() {
	defer log.Printf("<logpipe> finishing Close call\n")

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
	defer log.Printf("<logpipe> closing output channel\n")

	for {
		select {
		case t, ok := <-b.inchan:
			if !ok {
				return
			}
			log.Printf("<logpipe> type:%v   value:%v\n", reflect.TypeOf(t), t)
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
	log.Printf("<logpipe> created\n")

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func NewWithPipeline[T any](p pipelines.Pipeliner[T]) *LogPipe[T] {
	r := NewWithChannel(p.PipelineChan())
	r.pl = p
	log.Printf("<logpipe> pipeline set\n")
	return r
}

func New[T any]() *LogPipe[T] {
	return NewWithChannel(make(chan T, CHANSIZE))
}
