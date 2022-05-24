package retrypipe

import (
	"context"
	"log"
	"sync"

	"github.com/sterlingdevils/pipelines/pkg/containerpipe"
)

type Contextable interface {
	Context() context.Context
}

type Retryable interface {
	containerpipe.Keyable[uint64]
	Contextable
}

type Retry struct {
	objin  chan Retryable
	objout chan Retryable
	ackin  chan uint64

	wg sync.WaitGroup

	ctx  context.Context
	can  context.CancelFunc
	once sync.Once

	retrycontainer *containerpipe.ContainerPipe[uint64, Retryable]
}

const (
	CHANSIZE = 10
)

// ObjIn
func (r *Retry) ObjIn() chan<- Retryable {
	return r.objin
}

// ObjOut
func (r *Retry) ObjOut() <-chan Retryable {
	return r.objout
}

// AckIn
func (r *Retry) AckIn() chan<- uint64 {
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
func (r *Retry) checksendout(o Retryable) {
	defer recoverFromClosedChan()

	// Check if context expired, if so just drop it
	if o.Context().Err() != nil {
		return
	}

	// Send to output channel
	select {
	case <-o.Context().Done():
	case r.objout <- o:
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
func (r *Retry) mainloop() {
	defer r.wg.Done()

	for {
		select {
		case o := <-r.objin:
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
func (r *Retry) Close() {
	r.can()
	r.once.Do(func() {
		close(r.objout)
	})

	// close the retry container
	r.retrycontainer.Close()
}

// New
func New() (*Retry, error) {
	c, cancel := context.WithCancel(context.Background())
	oin := make(chan Retryable, CHANSIZE)
	oout := make(chan Retryable, CHANSIZE)
	ain := make(chan uint64, CHANSIZE)

	r := Retry{objin: oin, objout: oout, ackin: ain, ctx: c, can: cancel}

	// Create a retry container
	var err error
	r.retrycontainer, err = containerpipe.New[uint64, Retryable]()
	if err != nil {
		return nil, err
	}

	r.wg.Add(1)
	go r.mainloop()

	return &r, nil
}
