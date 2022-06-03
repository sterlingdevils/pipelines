package pipelines

import (
	"context"
	"sync"
)

type NullConsumePipe[T any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan chan T

	pl Pipeline[T]
	wg *sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b NullConsumePipe[T]) InChan() chan<- T {
	return b.inchan
}

// Close
func (b *NullConsumePipe[_]) Close() {
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
func (b *NullConsumePipe[_]) mainloop() {
	defer b.wg.Done()

	for {
		select {
		case _, ok := <-b.inchan:
			if !ok {
				return
			}
		case <-b.ctx.Done():
			return
		}
	}
}

func (NullConsumePipe[T]) NewWithChannel(in chan T) *NullConsumePipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := NullConsumePipe[T]{ctx: con, can: cancel, wg: new(sync.WaitGroup),
		inchan: in}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (n NullConsumePipe[T]) NewWithPipeline(p Pipeline[T]) *NullConsumePipe[T] {
	r := n.NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func (n NullConsumePipe[T]) New() *NullConsumePipe[T] {
	return n.NewWithChannel(make(chan T, CHANSIZE))
}
