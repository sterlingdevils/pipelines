package pipelines

import (
	"context"
	"log"
	"reflect"
	"sync"
)

type LogPipe[T any] struct {
	name string
	ctx  context.Context
	can  context.CancelFunc

	inchan  chan T
	outchan chan T

	pl Pipeline[T]
	wg *sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b LogPipe[T]) InChan() chan<- T {
	return b.inchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b LogPipe[T]) OutChan() <-chan T {
	return b.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b LogPipe[T]) PipelineChan() chan T {
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

func (LogPipe[T]) NewWithChannel(name string, in chan T) *LogPipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := LogPipe[T]{name: name,
		ctx: con, can: cancel, wg: new(sync.WaitGroup),
		inchan: in, outchan: make(chan T, CHANSIZE)}
	log.Printf("<logpipe %v> created\n", name)

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (b LogPipe[T]) NewWithPipeline(name string, p Pipeline[T]) *LogPipe[T] {
	r := b.NewWithChannel(name, p.PipelineChan())
	r.pl = p
	log.Printf("<logpipe %v> pipeline set\n", name)
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func (b LogPipe[T]) New(name string) *LogPipe[T] {
	return b.NewWithChannel(name, make(chan T, CHANSIZE))
}
