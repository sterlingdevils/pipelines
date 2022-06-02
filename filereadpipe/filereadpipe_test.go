package filereadpipe_test

import (
	"fmt"
	"log"
	"os"

	"github.com/sterlingdevils/pipelines/filereadpipe"
)

func Example_fileconsume() {
	err := os.WriteFile("/tmp/dat1", []byte("This is a test"), 0666)
	if err != nil {
		log.Fatalln(err)
	}

	fp := filereadpipe.New[filereadpipe.File]()

	fp.InChan() <- "/tmp/dat1"

	fmt.Println(string((<-fp.OutChan()).Data()))

	fp.Close()

	// Output:
	// This is a test
}
