package pipelines

import (
	"context"
	"errors"
	"sync"
)

type BufferPipe[T any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan T
	outchan chan T

	pl Pipeline[T]
	wg *sync.WaitGroup
}

// InChan
func (b BufferPipe[T]) InChan() chan<- T {
	return b.inchan
}

// OutChan
func (b BufferPipe[T]) OutChan() <-chan T {
	return b.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b BufferPipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *BufferPipe[_]) Close() {
	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()

	// Wait for us to be done
	b.wg.Wait()
}

// mainloop, read from in channel and write to out channel safely
// exit when our context is closed
func (b *BufferPipe[_]) mainloop() {
	defer b.wg.Done()
	defer close(b.outchan)

	for {
		select {
		case t := <-b.inchan:
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

func (BufferPipe[T]) NewWithChannel(size int, in chan T) (*BufferPipe[T], error) {
	if size < 1 {
		return nil, errors.New("buffer size must be >= 1")
	}

	con, cancel := context.WithCancel(context.Background())

	r := BufferPipe[T]{
		ctx:     con,
		can:     cancel,
		wg:      new(sync.WaitGroup),
		inchan:  in,
		outchan: make(chan T, size)}

	r.wg.Add(1)
	go r.mainloop()

	return &r, nil
}

func (b BufferPipe[T]) NewWithPipeline(size int, p Pipeline[T]) (*BufferPipe[T], error) {
	r, err := b.NewWithChannel(size, p.PipelineChan())
	if err != nil {
		return nil, err
	}

	r.pl = p

	return r, nil
}

func (BufferPipe[T]) New(size int) (*BufferPipe[T], error) {
	if size < 1 {
		return nil, errors.New("buffer size must be >= 1")
	}

	con, cancel := context.WithCancel(context.Background())
	r := BufferPipe[T]{
		ctx:     con,
		can:     cancel,
		wg:      new(sync.WaitGroup),
		inchan:  make(chan T, size),
		outchan: make(chan T, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r, nil
}
