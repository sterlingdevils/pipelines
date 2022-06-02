package filewriterpipe

import (
	"context"
	"os"
	"sync"

	"github.com/sterlingdevils/pipelines"
)

const (
	CHANSIZE = 0
)

type FileWriterPipe struct {
	ctx context.Context
	can context.CancelFunc

	inchan chan pipelines.FileNamerDataer

	pl pipelines.Pipeline[pipelines.FileNamerDataer]
	wg *sync.WaitGroup
}

// InChan returns a write only channel that the incomming packets will be read from
func (b FileWriterPipe) InChan() chan<- pipelines.FileNamerDataer {
	return b.inchan
}

// Close
func (b *FileWriterPipe) Close() {
	// If we pipelined then call Close the input pipeline
	if b.pl != nil {
		b.pl.Close()
	}

	// Cancel our context
	b.can()

	// Wait for us to be done
	b.wg.Wait()
}

func (b *FileWriterPipe) writefile(t pipelines.FileNamerDataer) {
	tmpName := "." + t.FileName()
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

	os.Rename(tmpName, t.FileName())
}

// mainloop, read from in channel and write to out channel safely, write the item
// exit when our context is closed
func (b *FileWriterPipe) mainloop() {
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

func NewWithChannel(in chan pipelines.FileNamerDataer) *FileWriterPipe {
	con, cancel := context.WithCancel(context.Background())
	r := FileWriterPipe{ctx: con, can: cancel, inchan: in, wg: new(sync.WaitGroup)}

	r.wg.Add(1)
	go r.mainloop()

	return &r
}

func NewWithPipeline(p pipelines.Pipeline[pipelines.FileNamerDataer]) *FileWriterPipe {
	r := NewWithChannel(p.PipelineChan())
	r.pl = p
	return r
}

// New creates a new FileWriterPipe
func New() *FileWriterPipe {
	return NewWithChannel(make(chan pipelines.FileNamerDataer, CHANSIZE))
}
