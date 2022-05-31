package filedump_test

import (
	"os"

	"github.com/sterlingdevils/pipelines/filedump"

	"github.com/sterlingdevils/pipelines/udppipe"
)

type DataHolder struct {
	data []byte
}

func (d DataHolder) Data() []byte {
	return d.data
}

func Example() {
	os.Chdir("/tmp")
	fd := filedump.New()

	// Send a Packet
	fd.InChan() <- udppipe.Packet{DataSlice: []byte("Hello, World!")}
	fd.InChan() <- udppipe.Packet{DataSlice: []byte("Gimme Jimmy")}
	fd.InChan() <- udppipe.Packet{DataSlice: []byte("See what happens with special characters\nOn this line")}

	fd.InChan() <- DataHolder{data: []byte("This is another type of input")}

	fd.Close()
	// Output:
	//
}
