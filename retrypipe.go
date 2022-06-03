package pipelines

import (
	"context"
	"sync"
	"time"
)

type Retryable[K comparable] interface {
	Keyer[K]
}

const (
	RETRYTIME  = 3 * time.Second
	EXPIRETIME = 30 * time.Second
)

type RetryPipe[K comparable, T Retryable[K]] struct {
	inchan  chan T
	outchan chan T
	ackin   chan K

	wg *sync.WaitGroup

	ctx context.Context
	can context.CancelFunc

	// Next one to retry
	nextone *RetryThing[K, T]

	retrycontainer *ContainerPipe[K, RetryThing[K, T]]

	pl Pipeline[T]

	RetryTime  time.Duration
	ExpireTime time.Duration
}

//
func (r RetryPipe[_, T]) InChan() chan<- T {
	return r.inchan
}

//
func (r RetryPipe[_, T]) OutChan() <-chan T {
	return r.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (r RetryPipe[_, T]) PipelineChan() chan T {
	return r.outchan
}

// AckIn
func (r RetryPipe[K, _]) AckIn() chan<- K {
	return r.ackin
}

// SetAckIn
func (r *RetryPipe[K, _]) SetAckIn(c chan K) {
	r.ackin = c
}

// chcecksendout do a safe write to the output channel
func (r RetryPipe[K, T]) retry(o *RetryThing[K, T]) {
	defer recoverFromClosedChan()

	// Check if we are expired
	if time.Since(o.Created()) > r.ExpireTime {
		return
	}

	// Send to output channel
	select {
	case r.outchan <- o.Thing():
	case <-r.ctx.Done():
		return
	}

	// Update Retry Time
	o.LastRetry = time.Now()

	select {
	case r.retrycontainer.InChan() <- *o:
	case <-r.ctx.Done():
		return
	}
}

// chcecksendout do a safe write to the output channel
func (r RetryPipe[K, T]) sendAndRetry(o T) {
	// Create new retry thing as this is the first time we have seen this
	rt := RetryThing[K, T]{}.New(o.Key(), o)

	// Now Send it
	r.retry(rt)
}

func minDuration(a, b time.Duration) time.Duration {
	if a <= b {
		return a
	}
	return b
}

// mainloop
func (r *RetryPipe[_, _]) mainloop() {
	defer r.wg.Done()
	defer close(r.outchan)

	for {

		if r.nextone == nil {
			select {
			case o, ok := <-r.inchan:
				if !ok {
					return
				}
				r.sendAndRetry(o)
			case a, ok := <-r.ackin:
				if !ok {
					return
				}
				r.retrycontainer.DelChan() <- a
			case o := <-r.retrycontainer.OutChan():
				r.nextone = &o
			case <-r.ctx.Done():
				return
			}
		} else {
			// So we have one to retry
			delay := minDuration(r.ExpireTime-time.Since(r.nextone.created), r.RetryTime-time.Since(r.nextone.LastRetry))
			retry := time.After(delay)

			select {
			// Check if the current retry one nees to be send
			case <-retry:
				r.retry(r.nextone)
				r.nextone = nil

			// Check for new incomming
			case o, ok := <-r.inchan:
				if !ok {
					return
				}
				r.sendAndRetry(o)

			// Check for Acks
			case a, ok := <-r.ackin:
				if !ok {
					return
				}
				// Check if our current retry waiting is the acked
				if a == r.nextone.Key() {
					r.nextone = nil
				}
				r.retrycontainer.DelChan() <- a

			// Check for Closed context
			case <-r.ctx.Done():
				return
			}
		}
	}
}

// Close us
func (r *RetryPipe[_, _]) Close() {
	// If we pipelined then call Close the input pipeline
	if r.pl != nil {
		r.pl.Close()
	}

	r.can()

	// close the retry container
	r.retrycontainer.Close()

	// Wait until we are finished
	r.wg.Wait()
}

// New with input channel
func (RetryPipe[K, T]) NewWithChannel(in chan T) *RetryPipe[K, T] {
	c, cancel := context.WithCancel(context.Background())
	oin := in
	oout := make(chan T, CHANSIZE)
	ain := make(chan K, CHANSIZE)

	r := RetryPipe[K, T]{inchan: oin, outchan: oout, ackin: ain,
		ctx: c, can: cancel, wg: new(sync.WaitGroup),
		RetryTime: RETRYTIME, ExpireTime: EXPIRETIME}

	// Create a retry container
	r.retrycontainer = ContainerPipe[K, RetryThing[K, T]]{}.New()

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

// New with pipeline
func (r RetryPipe[K, T]) NewWithPipeline(p Pipeline[T]) *RetryPipe[K, T] {
	n := r.NewWithChannel(p.PipelineChan())
	n.pl = p
	return n
}

// New
func (r RetryPipe[K, T]) New() *RetryPipe[K, T] {
	return r.NewWithChannel(make(chan T, CHANSIZE))
}
