package pipelines_test

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sterlingdevils/pipelines"
)

type DataHold struct {
	data []byte
}

func (d DataHold) Data() []byte {
	return d.data
}

func readOut(wg *sync.WaitGroup, o chan string) {
	defer wg.Done()
	count := 0
	for n := range o {
		count++
		fmt.Println(count)
		_ = n
	}
}

func Example() {
	os.Chdir("/tmp")
	fd := pipelines.FileDump{}.New()
	ochan := make(chan string)
	fd.Outchan = &ochan

	var wg sync.WaitGroup

	wg.Add(1)
	go readOut(&wg, ochan)

	// Send a Packet
	fd.InChan() <- pipelines.Packet{DataSlice: []byte("Hello, World!")}
	fd.InChan() <- pipelines.Packet{DataSlice: []byte("Gimme Jimmy")}
	fd.InChan() <- pipelines.Packet{DataSlice: []byte("See what happens with special characters\nOn this line")}

	fd.InChan() <- DataHold{data: []byte("This is another type of input")}

	time.Sleep(1 * time.Second)

	fd.Close()
	close(ochan)

	wg.Wait()
	// Output:
	// 1
	// 2
	// 3
	// 4
}
