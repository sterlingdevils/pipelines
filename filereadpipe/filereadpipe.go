package filereadpipe

import (
	"context"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sterlingdevils/pipelines"
)

const (
	CHANSIZE = 0
)

type File struct {
	Reference string
	data      []byte
}

func (f File) Data() []byte {
	return f.data
}

type FileReadPipe struct {
	ctx context.Context
	can context.CancelFunc

	inchan  chan string
	outchan chan pipelines.Dataer

	pl pipelines.Pipeline[string]
	wg *sync.WaitGroup
}

// PipelineChan returns a R/W channel that is used for pipelining
func (f *FileReadPipe) InChan() chan<- string {
	return f.inchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (f *FileReadPipe) OutChan() <-chan pipelines.Dataer {
	return f.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (f *FileReadPipe) PipelineChan() chan pipelines.Dataer {
	return f.outchan
}

// Close
func (f *FileReadPipe) Close() {
	// If we pipelined then call Close the input pipeline
	if f.pl != nil {
		f.pl.Close()
	}

	// Cancel our context
	f.can()

	// Wait for us to be done
	f.wg.Wait()
}

func (f *FileReadPipe) consumeFile(t string) {
	dat, err := ioutil.ReadFile(t)
	if err != nil {
		return
	}
	select {
	case f.outchan <- File{Reference: t, data: dat}:
		os.Remove(t)
	case <-f.ctx.Done():
		return
	}
}

// mainloop, read from in channel and write to out channel safely, log the item
// exit when our context is closed
func (f *FileReadPipe) mainloop() {
	defer f.wg.Done()
	defer close(f.outchan)

	for {
		select {
		case t, ok := <-f.inchan:
			if !ok {
				return
			}
			f.consumeFile(t)
		case <-f.ctx.Done():
			return
		}
	}
}

func NewWithChannel(in chan string) *FileReadPipe {
	con, cancel := context.WithCancel(context.Background())
	r := FileReadPipe{ctx: con, can: cancel, wg: new(sync.WaitGroup), inchan: in, outchan: make(chan pipelines.Dataer, CHANSIZE)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func NewWithPipeline(p pipelines.Pipeline[string]) *FileReadPipe {
	r := NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new filereadpipe
func New[T any]() *FileReadPipe {
	return NewWithChannel(make(chan string, CHANSIZE))
}
