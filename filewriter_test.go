package pipelines_test

import (
	"os"

	"github.com/sterlingdevils/pipelines"
)

type FileHolder struct {
	filename string
	data     []byte
}

func (f FileHolder) FileName() string {
	return f.filename
}

func (f FileHolder) Data() []byte {
	return f.data
}

type FileHolder2 struct {
	filename  string
	data      []byte
	Arbitrary []int
}

func (f FileHolder2) FileName() string {
	return f.filename
}

func (f FileHolder2) Data() []byte {
	return f.data
}

func ExampleFileWriterPipe() {
	os.Chdir("/tmp")
	fd := pipelines.FileWriterPipe{}.New()

	// Send a file
	fd.InChan() <- FileHolder{filename: "a.out", data: []byte("This is one type of input")}

	// Send a file as another type
	fd.InChan() <- FileHolder2{filename: "b.out", data: []byte("This is another type of input")}

	fd.Close()
	// Output:
	//
}
