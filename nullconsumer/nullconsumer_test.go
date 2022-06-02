package nullconsumer_test

import (
	"fmt"

	"github.com/sterlingdevils/pipelines/nullconsumer"
)

func Example() {
	nc := nullconsumer.New[any]()

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
