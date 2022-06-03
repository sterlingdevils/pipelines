package pipelines_test

import (
	"os"

	"github.com/sterlingdevils/pipelines"
)

type DataHold struct {
	data []byte
}

func (d DataHold) Data() []byte {
	return d.data
}

func ExampleFileDump() {
	os.Chdir("/tmp")
	fd := pipelines.FileDump{}.New()

	// Send a Packet
	fd.InChan() <- pipelines.Packet{DataSlice: []byte("Hello, World!")}
	fd.InChan() <- pipelines.Packet{DataSlice: []byte("Gimme Jimmy")}
	fd.InChan() <- pipelines.Packet{DataSlice: []byte("See what happens with special characters\nOn this line")}

	fd.InChan() <- DataHold{data: []byte("This is another type of input")}

	fd.Close()
	// Output:
	//
}
