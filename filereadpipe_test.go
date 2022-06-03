package pipelines_test

import (
	"fmt"
	"log"
	"os"

	"github.com/sterlingdevils/pipelines"
)

func ExampleFileReadPipe_fileconsume() {
	err := os.WriteFile("/tmp/dat1", []byte("This is a test"), 0666)
	if err != nil {
		log.Fatalln(err)
	}

	fp := pipelines.FileReadPipe{}.New()

	fp.InChan() <- "/tmp/dat1"

	fmt.Println(string((<-fp.OutChan()).Data()))

	fp.Close()

	// Output:
	// This is a test
}
