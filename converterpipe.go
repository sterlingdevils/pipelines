package pipelines

import (
	"context"
	"sync"
)

type ConverterPipe[I any, O any] struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan I
	outchan chan O

	convert func(I) (O, error)

	pl Pipeline[I]
	wg *sync.WaitGroup
}

// InChan
func (c ConverterPipe[I, O]) InChan() chan<- I {
	return c.inchan
}

// OutChan
func (c ConverterPipe[I, O]) OutChan() <-chan O {
	return c.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (c ConverterPipe[_, O]) PipelineChan() chan O {
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

	// Wait for us to be done
	c.wg.Wait()
}

// mainloop, read from in channel and write to out channel safely
// exit when our context is closed
func (c *ConverterPipe[I, O]) mainloop() {
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

func (ConverterPipe[I, O]) NewWithChannel(in chan I, fun func(I) (O, error)) *ConverterPipe[I, O] {
	con, cancel := context.WithCancel(context.Background())

	r := ConverterPipe[I, O]{
		ctx:     con,
		can:     cancel,
		wg:      new(sync.WaitGroup),
		convert: fun,
		inchan:  in,
		outchan: make(chan O, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (c ConverterPipe[I, O]) NewWithPipeline(p Pipeline[I], fun func(I) (O, error)) *ConverterPipe[I, O] {
	r := c.NewWithChannel(p.PipelineChan(), fun)
	r.pl = p

	return r
}

func (c ConverterPipe[I, O]) New(fun func(I) (O, error)) *ConverterPipe[I, O] {
	return c.NewWithChannel(make(chan I, CHANSIZE), fun)
}
