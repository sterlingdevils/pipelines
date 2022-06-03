package pipelines_test

import (
	"fmt"

	"github.com/sterlingdevils/pipelines"
)

func ExampleGeneratorPipe_increment() {
	i := 0
	gen := pipelines.GeneratorPipe[int]{}.New(
		func() int {
			i++
			return i
		})

	for j := 0; j < 10; j++ {
		o := <-gen.OutChan()
		fmt.Println(o)
	}

	gen.Close()
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
	// 10
}
