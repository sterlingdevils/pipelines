package retrypipe

import (
	"context"
	"log"
	"sync"

	"github.com/sterlingdevils/pipelines/pkg/containerpipe"
	"github.com/sterlingdevils/pipelines/pkg/pipeline"
)

type Contextable interface {
	Context() context.Context
}

type Retryable[K comparable] interface {
	containerpipe.Keyable[K]
	Contextable
}

type Retry[K comparable, T Retryable[K]] struct {
	inchan  chan T
	outchan chan T
	ackin   chan K

	wg sync.WaitGroup

	ctx  context.Context
	can  context.CancelFunc
	once sync.Once

	retrycontainer *containerpipe.ContainerPipe[K, T]

	pl pipeline.Pipelineable[T]
}

const (
	CHANSIZE = 10
)

// ObjIn
func (r *Retry[_, T]) InChan() chan<- T {
	return r.inchan
}

// ObjOut
func (r *Retry[_, T]) OutChan() <-chan T {
	return r.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (r *Retry[_, T]) PipelineChan() chan T {
	return r.outchan
}

// AckIn
func (r *Retry[K, _]) AckIn() chan<- K {
	return r.ackin
}

// RecoverFromClosedChan is used when it is OK if the channel is closed we are writing on
// This is not great using the string compare but the go runtime uses a generic error so we
// can't trap this any other way.
func recoverFromClosedChan() {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok && e.Error() == "send on closed channel" {
			log.Println("might be recovery from closed channel. not a problem: ", e)
		} else {
			panic(r)
		}
	}
}

// chcecksendout do a safe write to the output channel
func (r *Retry[K, T]) checksendout(o T) {
	defer recoverFromClosedChan()

	// Check if context expired, if so just drop it
	if o.Context().Err() != nil {
		return
	}

	// Send to output channel
	select {
	case <-o.Context().Done():
	case r.outchan <- o:
	case <-r.ctx.Done():
		return
	}

	// Send to retry channel
	select {
	case <-o.Context().Done():
	case r.retrycontainer.InChan() <- o:
	case <-r.ctx.Done():
		return
	}
}

// mainloop
func (r *Retry[_, _]) mainloop() {
	defer r.wg.Done()

	for {
		select {
		case o := <-r.inchan:
			r.checksendout(o)
		case a := <-r.ackin:
			r.retrycontainer.DelChan() <- a
		case o := <-r.retrycontainer.OutChan():
			_ = o
		case <-r.ctx.Done():
			return
		}
	}
}

// Close us
func (r *Retry[_, _]) Close() {
	// If we pipelined then call Close the input pipeline
	if r.pl != nil {
		r.pl.Close()
	}

	r.can()
	r.once.Do(func() {
		close(r.outchan)
	})

	// close the retry container
	r.retrycontainer.Close()
}

// New
func New[K comparable, T Retryable[K]]() (*Retry[K, T], error) {
	c, cancel := context.WithCancel(context.Background())
	oin := make(chan T, CHANSIZE)
	oout := make(chan T, CHANSIZE)
	ain := make(chan K, CHANSIZE)

	r := Retry[K, T]{inchan: oin, outchan: oout, ackin: ain, ctx: c, can: cancel}

	// Create a retry container
	var err error
	r.retrycontainer, err = containerpipe.New[K, T]()
	if err != nil {
		return nil, err
	}

	r.wg.Add(1)
	go r.mainloop()

	return &r, nil
}
