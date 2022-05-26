package converterpipe

import (
	"context"

	"github.com/sterlingdevils/pipelines/pkg/pipeline"
)

const (
	CHANSIZE = 0
)

type ConverterPipe[I any, O any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan I
	outchan chan O

	convert func(I) O

	pl pipeline.Pipelineable[I]
}

// InChan
func (c *ConverterPipe[I, O]) InChan() chan<- I {
	return c.inchan
}

// OutChan
func (c *ConverterPipe[I, O]) OutChan() <-chan O {
	return c.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (c *ConverterPipe[_, O]) PipelineChan() chan O {
	return c.outchan
}

// Close
func (c *ConverterPipe[_, _]) Close() {
	// If we pipelined then call Close the input pipeline
	if c.pl != nil {
		c.pl.Close()
	}

	// Cancel our context
	c.can()
}

// mainloop, read from in channel and write to out channel safely
// exit when our context is closed
func (c *ConverterPipe[I, O]) mainloop() {
	defer close(c.outchan)

	for {
		select {
		case t := <-c.inchan:
			v := c.convert(t)
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

func NewWithChannel[I, O any](fun func(I) O, in chan I) *ConverterPipe[I, O] {
	con, cancel := context.WithCancel(context.Background())

	r := ConverterPipe[I, O]{
		ctx:     con,
		can:     cancel,
		convert: fun,
		inchan:  in,
		outchan: make(chan O, CHANSIZE)}

	go r.mainloop()

	return &r
}

func NewWithPipeline[I, O any](fun func(I) O, p pipeline.Pipelineable[I]) *ConverterPipe[I, O] {
	r := NewWithChannel(fun, p.PipelineChan())
	r.pl = p

	return r
}

func New[I, O any](fun func(I) O) *ConverterPipe[I, O] {
	return NewWithChannel(fun, make(chan I, CHANSIZE))
}
