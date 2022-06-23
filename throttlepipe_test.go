package pipelines_test

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/sterlingdevils/pipelines"
)

func writePipe(p *pipelines.ThrottlePipe[int]) {
	for i := 0; i < 10; i++ {
		//fmt.Printf("Writing: %v\n", i)
		p.InChan() <- i
	}
}

func ExampleThrottlePipe_New() {
	log.Printf("Starting test\n")
	throt := pipelines.ThrottlePipe[int]{}.New()
	conv := pipelines.ConverterPipe[int, int]{}.NewWithPipeline(throt,
		func(i int) (int, error) {
			fmt.Printf("Reading: %v\n", i)
			return 0, errors.New("This is not an error")
		})
	nullpipe := pipelines.NullConsumePipe[int]{}.NewWithPipeline(conv)
	defer nullpipe.Close()

	go writePipe(throt)

	time.Sleep(2 * time.Second)
	throt.SetTok() <- 5
	throt.AddTok() <- 2

	time.Sleep(2 * time.Second)
	throt.SetTok() <- 2

	time.Sleep(1 * time.Second)
	throt.SetTok() <- 3

	time.Sleep(1 * time.Second)
	//Output:
	// Reading: 0
	// Reading: 1
	// Reading: 2
	// Reading: 3
	// Reading: 4
	// Reading: 5
	// Reading: 6
	// Reading: 7
	// Reading: 8
	// Reading: 9
}

func ExampleThrottlePipe2() {
	log.Printf("Starting test\n")
	throt := pipelines.ThrottlePipe[int]{}.New()
	conv := pipelines.ConverterPipe[int, int]{}.NewWithPipeline(throt,
		func(i int) (int, error) {
			fmt.Printf("Reading: %v\n", i)
			return 0, errors.New("This is not an error")
		})
	nullpipe := pipelines.NullConsumePipe[int]{}.NewWithPipeline(conv)
	defer nullpipe.Close()

	go writePipe(throt)

	for i := 1; i < 6; i++ {
		time.Sleep(1 * time.Second)
		throt.AddTok() <- uint64(i)
	}

	time.Sleep(1 * time.Second)
	//Output:
	// Reading: 0
	// Reading: 1
	// Reading: 2
	// Reading: 3
	// Reading: 4
	// Reading: 5
	// Reading: 6
	// Reading: 7
	// Reading: 8
	// Reading: 9
}
