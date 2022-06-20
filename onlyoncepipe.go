package pipelines

import (
	"context"
	"sync"
	"time"
)

type OnlyOncePipe[T comparable] struct {
	smap   map[T]time.Time
	gctime time.Duration
	frtime time.Duration

	ctx context.Context
	can context.CancelFunc

	inchan  chan T
	outchan chan T

	pl Pipeline[T]
	wg *sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b OnlyOncePipe[T]) InChan() chan<- T {
	return b.inchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b OnlyOncePipe[T]) OutChan() <-chan T {
	return b.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (b OnlyOncePipe[T]) PipelineChan() chan T {
	return b.outchan
}

// Close
func (b *OnlyOncePipe[_]) Close() {
	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()

	// Wait for us to be done
	b.wg.Wait()
}

// mainloop, read from in channel and write to out channel safely,
// add it to the map if it isn't already there. Exit when our context is closed
func (b *OnlyOncePipe[_]) mainloop() {
	defer b.wg.Done()
	defer close(b.outchan)

	ticker := time.NewTicker(b.gctime)

	for {
		select {
		case <-ticker.C:
			for m, val := range b.smap {
				if val.Add(time.Duration(b.frtime.Seconds())).After(time.Now()) {
					delete(b.smap, m)
				}
			}
		case t, ok := <-b.inchan:
			if !ok {
				return
			}
			_, ok = b.smap[t]
			if !ok {
				b.smap[t] = time.Now()
				select {
				case b.outchan <- t:
				case <-b.ctx.Done():
					return
				}
			}
		case <-b.ctx.Done():
			return
		}
	}
}

func (OnlyOncePipe[T]) NewWithChannel(gc time.Duration, fr time.Duration, in chan T) *OnlyOncePipe[T] {
	con, cancel := context.WithCancel(context.Background())
	r := OnlyOncePipe[T]{smap: make(map[T]time.Time), gctime: gc, frtime: fr,
		ctx: con, can: cancel, wg: new(sync.WaitGroup),
		inchan: in, outchan: make(chan T, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (b OnlyOncePipe[T]) NewWithPipeline(gc time.Duration, fr time.Duration, p Pipeline[T]) *OnlyOncePipe[T] {
	r := b.NewWithChannel(gc, fr, p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func (b OnlyOncePipe[T]) New(gc time.Duration, fr time.Duration) *OnlyOncePipe[T] {
	return b.NewWithChannel(gc, fr, make(chan T, CHANSIZE))
}
