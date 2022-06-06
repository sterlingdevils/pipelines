package pipelines

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type FileDump struct {
	received uint64

	ctx context.Context
	can context.CancelFunc

	inchan  chan Dataer
	Outchan *chan string

	pl Pipeline[Dataer]
	wg *sync.WaitGroup

	Metricfunc func(name string, val int)
}

func (b FileDump) SetMetric(name string, val int) {
	if b.Metricfunc == nil {
		return
	}

	b.Metricfunc(name, val)
}

// InChan returns a write only channel that the incomming packets will be read from
func (b FileDump) InChan() chan<- Dataer {
	return b.inchan
}

// Close
func (b *FileDump) Close() {
	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()

	// Wait for us to be done
	b.wg.Wait()
}

func (b *FileDump) writefile(t Dataer) (string, error) {
	name := strconv.FormatInt(time.Now().Unix(), 10) + "." + fmt.Sprintf("%06d", b.received)
	tmpName := "." + name
	tmpFd, err := os.Create(tmpName)
	if err != nil {
		return "", errors.New("Unable to create file")
	}

	_, err = tmpFd.Write(t.Data())
	if err != nil {
		tmpFd.Close()
		return "", errors.New("Unable to write to file")
	}
	tmpFd.Close()

	os.Rename(tmpName, name)
	b.received++

	b.SetMetric("filecount", int(b.received))
	return name, nil
}

// Write to a file and put the name on the output buffer
func (b *FileDump) writeandRespond(t Dataer) {
	f, err := b.writefile(t)
	if err != nil {
		return
	}

	if b.Outchan == nil {
		return
	}

	select {
	case *b.Outchan <- f:
	case <-b.ctx.Done():
	}
}

// mainloop, read from in channel and write to out channel safely, write the item
// exit when our context is closed
func (b *FileDump) mainloop() {
	defer b.wg.Done()

	for {
		select {
		case t, ok := <-b.inchan:
			if !ok {
				return
			}
			b.writeandRespond(t)
		case <-b.ctx.Done():
			return
		}
	}
}

func (FileDump) NewWithChannel(in chan Dataer) *FileDump {
	con, cancel := context.WithCancel(context.Background())
	r := FileDump{received: 0, ctx: con, can: cancel, inchan: in, wg: new(sync.WaitGroup)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func (f FileDump) NewWithPipeline(p Pipeline[Dataer]) *FileDump {
	r := f.NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new FileDump
func (f FileDump) New() *FileDump {
	return f.NewWithChannel(make(chan Dataer, CHANSIZE))
}
