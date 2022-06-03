/*
  package chanbaedcontainer implements an ordered container with only channels as the API

  The input channel is used to add things to the container.
  The output channel will contain the head of the container when read
  The delete channel is used to remove things from the container before the are read out of
  the output channel.

  This uses Go 1.18 generics,  Things must impement the Indexable interface:
    has a method to return a comparable key

  Things come in and go out the channels in order.  Things can be removed while in the container
  by passing their key to the delete channel
*/
package pipelines

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
)

type ContainerPipe[K comparable, T Keyer[K]] struct {
	// We use map to hold the thing, and an ordered list of keys
	tmap  map[K]T
	tlist *list.List

	inchan  chan T
	outchan chan T
	delchan chan K

	ctx context.Context
	can context.CancelFunc

	// Holds the current thing we are trying to send
	onetosend *T

	approxSize int32

	pl Pipeline[T]
	wg *sync.WaitGroup
}

func (c *ContainerPipe[_, T]) addT(thing T) {
	k := thing.Key()
	if _, b := c.tmap[k]; b {
		return
	}

	c.tlist.PushBack(k)
	c.tmap[k] = thing
}

func (c *ContainerPipe[K, _]) delK(index K) {
	// If we have one to delete, check if its the one we are waiting to send
	if c.onetosend != nil {
		if (*c.onetosend).Key() == index {
			c.onetosend = nil
			return
		}
	}

	// Not onetosend so check in the continer
	for curr := c.tlist.Front(); curr != nil; curr = curr.Next() {
		val := curr.Value.(K)
		if val == index {
			c.tlist.Remove(curr)
			delete(c.tmap, index)
			break
		}
	}
}

// grab the head of the container or nil if we are empty
func (c *ContainerPipe[K, T]) pop() *T {
	if len(c.tmap) == 0 {
		return nil
	}

	e := c.tlist.Front()
	k := e.Value.(K)
	t := c.tmap[k]
	c.delK(k)
	return &t
}

// ApproxSize returns something close to the number of items in the container, maybe.
// Only updated at the start of each mainloop
func (c ContainerPipe[_, _]) ApproxSize() int32 {
	return atomic.LoadInt32(&c.approxSize)
}

// InChan
func (c ContainerPipe[_, T]) InChan() chan<- T {
	return c.inchan
}

// DelChan
func (c ContainerPipe[K, _]) DelChan() chan<- K {
	return c.delchan
}

// OutChan
func (c ContainerPipe[_, T]) OutChan() <-chan T {
	return c.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (c ContainerPipe[_, T]) PipelineChan() chan T {
	return c.outchan
}

// Close the ChanBasedContainer
func (c *ContainerPipe[_, _]) Close() {
	// If we pipelined then call Close the input pipeline
	if c.pl != nil {
		c.pl.Close()
	}

	// Cancel our context
	c.can()

	// Wait for us to be done
	c.wg.Wait()
}

// mainloop
// If the container is empty, only listen for
func (c *ContainerPipe[_, T]) mainloop() {
	defer c.wg.Done()
	defer close(c.outchan)
	defer recoverFromClosedChan()

	for {
		// Check if we have one ready to send
		if c.onetosend == nil {
			c.onetosend = c.pop() // pop will return nil if one is not ready
		}

		if c.onetosend == nil {
			// Save the current size
			atomic.StoreInt32(&c.approxSize, int32(len(c.tmap)))
			// None to send so don't select on output channel
			select {
			case t := <-c.inchan:

				c.addT(t)
			case k := <-c.delchan:
				c.delK(k)
			case <-c.ctx.Done():
				return
			}
		} else {
			// Save the current size
			atomic.StoreInt32(&c.approxSize, int32(len(c.tmap))+1)

			// We have one to send so select on output channel
			select {
			case c.outchan <- *c.onetosend:
				// Now that we sent it, clean onetosend so we get the next one
				c.onetosend = nil
			case t := <-c.inchan:
				c.addT(t)
			case k := <-c.delchan:
				c.delK(k)
			case <-c.ctx.Done():
				return
			}
		}
	}
}

func (ContainerPipe[K, T]) NewWithChan(in chan T) *ContainerPipe[K, T] {
	con, cancel := context.WithCancel(context.Background())
	r := ContainerPipe[K, T]{
		tmap:    make(map[K]T),
		tlist:   list.New(),
		inchan:  in,
		outchan: make(chan T, CHANSIZE),
		delchan: make(chan K, CHANSIZE),
		wg:      new(sync.WaitGroup),
		ctx:     con,
		can:     cancel}

	r.wg.Add(1)
	go r.mainloop()
	return &r
}

func (c ContainerPipe[K, T]) NewWithPipeline(p Pipeline[T]) *ContainerPipe[K, T] {
	r := c.NewWithChan(p.PipelineChan())

	// save pipeline
	r.pl = p

	return r
}

// New returns a reference to a a container or error if there was a problem
// for performance T should be a pointer
func (c ContainerPipe[K, T]) New() *ContainerPipe[K, T] {
	r := c.NewWithChan(make(chan T, CHANSIZE))
	return r
}
