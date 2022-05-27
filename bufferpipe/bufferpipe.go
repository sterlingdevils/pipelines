package bufferpipe

import (
	"context"
	"errors"
	"sync"

	"github.com/sterlingdevils/pipelines"
)

const (
	CHANSIZE = 0
)

type BufferPipe[T any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan T
	outchan chan T

	pl pipelines.Pipeliner[T]
	wg sync.WaitGroup
}

// InChan
func (b *BufferPipe[T]) InChan() chan<- T {
	return b.inchan
}

// OutChan
func (b *BufferPipe[T]) OutChan() <-chan T {
	return b.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b *BufferPipe[T]) PipelineChan() chan T {
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

func NewWithChannel[T any](size int, in chan T) (*BufferPipe[T], error) {
	if size < 1 {
		return nil, errors.New("buffer size must be >= 1")
	}

	con, cancel := context.WithCancel(context.Background())

	r := BufferPipe[T]{
		ctx:     con,
		can:     cancel,
		inchan:  in,
		outchan: make(chan T, size)}

	r.wg.Add(1)
	go r.mainloop()

	return &r, nil
}

func NewWithPipeline[T any](size int, p pipelines.Pipeliner[T]) (*BufferPipe[T], error) {
	r, err := NewWithChannel(size, p.PipelineChan())
	if err != nil {
		return nil, err
	}

	r.pl = p

	return r, nil
}

func New[T any](size int) (*BufferPipe[T], error) {
	if size < 1 {
		return nil, errors.New("buffer size must be >= 1")
	}

	con, cancel := context.WithCancel(context.Background())
	r := BufferPipe[T]{
		ctx:     con,
		can:     cancel,
		inchan:  make(chan T, size),
		outchan: make(chan T, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r, nil
}
