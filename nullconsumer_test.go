package pipelines_test

import (
	"fmt"

	"github.com/sterlingdevils/pipelines"
)

func ExampleNullConsumePipe() {
	nc := pipelines.NullConsumePipe[any]{}.New()

	inchan := nc.InChan()

	inchan <- 1
	inchan <- 2
	inchan <- 3
	inchan <- 4
	inchan <- 5

	nc.Close()

	fmt.Println("reached end")

	// Output:
	// reached end
}
