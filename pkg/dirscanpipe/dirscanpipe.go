package dirscanpipe

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DirScan holds a directry to scan and a Channel to put the filenames onto
type DirScan struct {
	Dir      string
	ScanTime time.Duration

	outchan chan string

	ctx  context.Context
	can  context.CancelFunc
	once sync.Once
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

// scanDir
// This is Blocking on the channel write
func (d *DirScan) scanDir() error {
	// This is to recover if we write to a closed channel, that is not a problem so recover from a panic
	defer recoverFromClosedChan()

	entries, err := os.ReadDir(d.Dir)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(d.Dir)
	if err != nil {
		return err
	}

	for _, f := range entries {
		// if it is not a directory and is not .prefixed
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			select {
			case d.outchan <- filepath.Join(path, f.Name()):
			case <-d.ctx.Done():
				return nil
			}
		}
	}
	return nil
}

// loop until we receive a stop on the run channel
func (d *DirScan) loop(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		d.scanDir()
		select {
		case <-time.After(d.ScanTime):
		case <-d.ctx.Done():
			return
		}
	}
}

// -----  Public Methods
// InChan, used for pipelineable
func (d *DirScan) InChan() chan<- string {
	return nil
}

// OutChan returns the output Channel as ReadOnly
func (d *DirScan) OutChan() <-chan string {
	return d.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (d *DirScan) PipelineChan() chan string {
	return d.outchan
}

// Close will close the data channel
func (d *DirScan) Close() {
	d.can()
	d.once.Do(func() {
		close(d.outchan)
	})
}

// New creates a new dir scanner and starts a scanning loop to send filenames to a channel
// Must pass a WaitGroup it as we create a go routine for the scanner
// As a writter we assume we own the channel we return, we will close it when our Close() is called
func New(wg *sync.WaitGroup, dir string, scantime time.Duration, chanSize int) (*DirScan, error) {
	if wg == nil {
		return nil, errors.New("must provide a valid waitgroup")
	}

	if fileInfo, err := os.Stat(dir); err != nil {
		return nil, errors.New("name not found: " + dir)
	} else {
		if !fileInfo.IsDir() {
			return nil, errors.New("name is not a directory: " + dir)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	d := DirScan{Dir: dir, outchan: make(chan string, chanSize), ScanTime: scantime, ctx: ctx, can: cancel}

	wg.Add(1)
	go d.loop(wg)

	return &d, nil
}
