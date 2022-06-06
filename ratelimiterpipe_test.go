package pipelines_test

import (
	"fmt"

	"github.com/sterlingdevils/pipelines"
)

type DataType string

type StringNode struct {
	data DataType
}

func (n StringNode) Size() int {
	return len(n.data)
}

func (n StringNode) Data() []byte {
	return []byte(n.data)
}

func ExampleNew() {
	_ = pipelines.RateLimiterPipe[StringNode]{}.New(1, 2)
	// Output:
	//
}

func Example_testsend() {
	n := StringNode{data: "potatoes"}
	r := pipelines.RateLimiterPipe[StringNode]{}.New(4, n.Size())
	r.InChan() <- n
	t := <-r.OutChan()
	fmt.Println(t.data)
	// Output:
	// potatoes
}

func Example_testsend2() {
	n := StringNode{data: "potatoes"}
	r := pipelines.RateLimiterPipe[StringNode]{}.New(1, n.Size())
	r.InChan() <- n
	t := <-r.OutChan()
	r.InChan() <- n
	t = <-r.OutChan()
	fmt.Println(t.data)
	// Output:
	// potatoes
}
