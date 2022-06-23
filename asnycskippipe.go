package pipelines

import (
	"context"
	"sync"
)

type AsyncSkipPipe[T any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan T
	outchan chan T

	pl Pipeline[T]
	wg *sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b AsyncSkipPipe[T]) InChan() chan<- T {
	return b.inchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b AsyncSkipPipe[T]) OutChan() <-chan T {
	return b.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b AsyncSkipPipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *AsyncSkipPipe[_]) Close() {
	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()

	// Wait for us to be done
	b.wg.Wait()
}

// mainloop, read from in channel and write to out channel if it is available
// exit when our context is closed
func (b *AsyncSkipPipe[_]) mainloop() {
	defer b.wg.Done()
	defer close(b.outchan)

	for {
		select {
		case t, ok := <-b.inchan:
			if !ok {
				return
			}
			select {
			case b.outchan <- t:
			case <-b.ctx.Done():
				return
			default:
			}
		case <-b.ctx.Done():
			return
		}
	}
}

func (AsyncSkipPipe[T]) NewWithChannel(in chan T) *AsyncSkipPipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := AsyncSkipPipe[T]{
		ctx: con, can: cancel, wg: new(sync.WaitGroup),
		inchan: in, outchan: make(chan T, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (b AsyncSkipPipe[T]) NewWithPipeline(p Pipeline[T]) *AsyncSkipPipe[T] {
	r := b.NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func (b AsyncSkipPipe[T]) New() *AsyncSkipPipe[T] {
	return b.NewWithChannel(make(chan T, CHANSIZE))
}
