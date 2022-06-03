package typeconverterpipe

import (
	"context"
	"errors"
	"sync"

	"github.com/sterlingdevils/pipelines"
)

const (
	CHANSIZE = 0
)

type TypeConverterPipe[I any, O any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan I
	outchan chan O

	pl pipelines.Pipeline[I]
	wg *sync.WaitGroup
}

// InChan
func (c TypeConverterPipe[I, O]) InChan() chan<- I {
	return c.inchan
}

// OutChan
func (c TypeConverterPipe[I, O]) OutChan() <-chan O {
	return c.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (c TypeConverterPipe[_, O]) PipelineChan() chan O {
	return c.outchan
}

// Close
func (c *TypeConverterPipe[_, _]) Close() {
	// If we pipelined then call Close the input pipeline
	if c.pl != nil {
		c.pl.Close()
	}

	// Cancel our context
	c.can()

	// Wait for us to be done
	c.wg.Wait()
}

func (c *TypeConverterPipe[I, O]) convert(i I) (O, error) {
	var p any = i
	v, ok := p.(O)
	if !ok {
		var e O
		return e, errors.New("cant convert I to O")
	}

	return v, nil
}

// mainloop, read from in channel and write to out channel safely
// exit when our context is closed
func (c *TypeConverterPipe[I, O]) mainloop() {
	defer c.wg.Done()
	defer close(c.outchan)

	for {
		select {
		case t, ok := <-c.inchan:
			if !ok {
				return
			}
			v, err := c.convert(t)
			if err != nil {
				break
			}
			select {
			case c.outchan <- v:
			case <-c.ctx.Done():
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func NewWithChannel[I, O any](in chan I) *TypeConverterPipe[I, O] {
	con, cancel := context.WithCancel(context.Background())

	r := TypeConverterPipe[I, O]{
		ctx:     con,
		can:     cancel,
		wg:      new(sync.WaitGroup),
		inchan:  in,
		outchan: make(chan O, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func NewWithPipeline[I, O any](p pipelines.Pipeline[I]) *TypeConverterPipe[I, O] {
	r := NewWithChannel[I, O](p.PipelineChan())
	r.pl = p

	return r
}

func New[I, O any]() *TypeConverterPipe[I, O] {
	return NewWithChannel[I, O](make(chan I, CHANSIZE))
}
