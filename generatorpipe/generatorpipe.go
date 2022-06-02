package generatorpipe

import (
	"context"
	"sync"

	"github.com/sterlingdevils/gobase/metrics"
)

const (
	CHANSIZE = 0
)

type GeneratorPipe[T any] struct {
	ctx context.Context
	can context.CancelFunc

	outchan chan T

	generate func() T

	wg *sync.WaitGroup

	Metricfunc func(metrics.MetricsProto)
}

func (g GeneratorPipe[_]) incMetric(name string) {
	if g.Metricfunc == nil {
		return
	}

	g.Metricfunc(metrics.MetricsProto{Name: name, Cmd: metrics.INC})
}

// OutChan
func (g GeneratorPipe[T]) OutChan() <-chan T {
	return g.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (g GeneratorPipe[T]) PipelineChan() chan T {
	return g.outchan
}

// Close
func (g *GeneratorPipe[T]) Close() {
	// Cancel our context
	g.can()

	// Wait for us to be done
	g.wg.Wait()
}

// mainloop, read from in channel and write to out channel safely
// exit when our context is closed
func (g *GeneratorPipe[T]) mainloop() {
	defer g.wg.Done()
	defer close(g.outchan)

	for {
		select {
		case g.outchan <- g.generate():
			g.incMetric("count")
		case <-g.ctx.Done():
			return
		}
	}
}

func New[T any](fun func() T) *GeneratorPipe[T] {
	con, cancel := context.WithCancel(context.Background())

	r := GeneratorPipe[T]{
		ctx:      con,
		can:      cancel,
		wg:       new(sync.WaitGroup),
		generate: fun,
		outchan:  make(chan T, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}
