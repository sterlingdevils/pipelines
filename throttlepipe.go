package pipelines

import (
	"context"
	"sync"
)

type ThrottlePipe[T any] struct {
	tokens uint64

	ctx context.Context
	can context.CancelFunc

	addTok  chan uint64
	setTok  chan uint64
	inchan  chan T
	outchan chan T

	pl Pipeline[T]
	wg *sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b ThrottlePipe[T]) InChan() chan<- T {
	return b.inchan
}

// Returns a write channel used for setting the number of tokens
func (b ThrottlePipe[T]) SetTok() chan<- uint64 {
	return b.setTok
}

// Returns a write channel used for adding to the number of tokens
func (b ThrottlePipe[T]) AddTok() chan<- uint64 {
	return b.addTok
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b ThrottlePipe[T]) OutChan() <-chan T {
	return b.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b ThrottlePipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *ThrottlePipe[_]) Close() {
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
func (b *ThrottlePipe[T]) mainloop() {
	defer b.wg.Done()
	defer close(b.outchan)

	for {
		// If token is available
		if b.tokens > 0 {
			select {
			case t, ok := <-b.inchan:
				if !ok {
					return
				}
				b.tokens--
				b.outchan <- t
			case t, ok := <-b.setTok:
				if !ok {
					return
				}
				b.tokens = t
			case t, ok := <-b.addTok:
				if !ok {
					return
				}
				b.tokens += t
			case <-b.ctx.Done():
				return
			}
		} else {
			// If tokens are not available
			select {
			case t, ok := <-b.setTok:
				if !ok {
					return
				}
				b.tokens = t
			case t, ok := <-b.addTok:
				if !ok {
					return
				}
				b.tokens += t
			case <-b.ctx.Done():
				return
			}
		}
	}
}

func (ThrottlePipe[T]) NewWithChannel(in chan T) *ThrottlePipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := ThrottlePipe[T]{tokens: 0,
		setTok: make(chan uint64), addTok: make(chan uint64),
		ctx: con, can: cancel, wg: new(sync.WaitGroup),
		inchan: in, outchan: make(chan T, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (b ThrottlePipe[T]) NewWithPipeline(p Pipeline[T]) *ThrottlePipe[T] {
	r := b.NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func (b ThrottlePipe[T]) New() *ThrottlePipe[T] {
	return b.NewWithChannel(make(chan T, CHANSIZE))
}
