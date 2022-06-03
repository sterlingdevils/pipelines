package pipelines_test

import (
	"fmt"

	"github.com/sterlingdevils/pipelines"
)

func ExampleLogPipe() {
	lg := pipelines.LogPipe[int]{}.New("example")
	lg.InChan() <- 5
	fmt.Println(<-lg.OutChan())
	// Output:
	// 5
}
