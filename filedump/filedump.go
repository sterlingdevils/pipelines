package filedump

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sterlingdevils/pipelines"
)

const (
	CHANSIZE = 0
)

type FileDump struct {
	received uint64

	ctx context.Context
	can context.CancelFunc

	inchan chan pipelines.Dataer

	pl pipelines.Pipeline[pipelines.Dataer]
	wg sync.WaitGroup
}

// InChan returns a write only channel that the incomming packets will be read from
func (b FileDump) InChan() chan<- pipelines.Dataer {
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

func (b *FileDump) writefile(t pipelines.Dataer) {
	name := strconv.FormatInt(time.Now().Unix(), 10) + "." + fmt.Sprintf("%06d", b.received)
	tmpName := "." + name
	tmpFd, err := os.Create(tmpName)
	if err != nil {
		return
	}

	_, err = tmpFd.Write(t.Data())
	if err != nil {
		tmpFd.Close()
		return
	}
	tmpFd.Close()

	os.Rename(tmpName, name)
	b.received++
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
			b.writefile(t)
		case <-b.ctx.Done():
			return
		}
	}
}

func NewWithChannel(in chan pipelines.Dataer) *FileDump {
	con, cancel := context.WithCancel(context.Background())
	r := FileDump{received: 0, ctx: con, can: cancel, inchan: in}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func NewWithPipeline(p pipelines.Pipeline[pipelines.Dataer]) *FileDump {
	r := NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new logger
// name is used to put unique label on each log
func New() *FileDump {
	return NewWithChannel(make(chan pipelines.Dataer, CHANSIZE))
}
