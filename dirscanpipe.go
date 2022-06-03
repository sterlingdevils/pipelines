package pipelines

import (
	"context"
	"errors"
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

	ctx context.Context
	can context.CancelFunc

	wg *sync.WaitGroup
}

// scanDir
// This is Blocking on the channel write
func (d DirScan) scanDir() error {
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
func (d *DirScan) mainloop() {
	defer d.wg.Done()
	defer close(d.outchan)

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
// OutChan returns the output Channel as ReadOnly
func (d DirScan) OutChan() <-chan string {
	return d.outchan
}

// PipelineChan returns a R/W channel that is used for pipelining
func (d DirScan) PipelineChan() chan string {
	return d.outchan
}

// Close will close the data channel
func (d *DirScan) Close() {
	d.can()

	// Wait for us to be done
	d.wg.Wait()
}

// New creates a new dir scanner and starts a scanning loop to send filenames to a channel
// Must pass a WaitGroup it as we create a go routine for the scanner
// As a writter we assume we own the channel we return, we will close it when our Close() is called
func (DirScan) New(dir string, scantime time.Duration, chanSize int) (*DirScan, error) {
	if fileInfo, err := os.Stat(dir); err != nil {
		return nil, errors.New("name not found: " + dir)
	} else {
		if !fileInfo.IsDir() {
			return nil, errors.New("name is not a directory: " + dir)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	d := DirScan{Dir: dir, outchan: make(chan string, chanSize), ScanTime: scantime, ctx: ctx, can: cancel, wg: new(sync.WaitGroup)}

	d.wg.Add(1)
	go d.mainloop()

	return &d, nil
}
