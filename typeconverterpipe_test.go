package pipelines_test

import (
	"fmt"
	"reflect"

	"github.com/sterlingdevils/pipelines"
)

type Iface interface {
	Boo() int
}

type tctA struct{}

func (a tctA) Boo() int {
	return 2
}

func ExampleTypeConverterPipe() {
	cvt := pipelines.TypeConverterPipe[tctA, Iface]{}.New()

	cvt.InChan() <- tctA{}
	var o Iface = <-cvt.OutChan()

	fmt.Println(o.Boo(), reflect.TypeOf(o))

	// Output:
	// 2 pipelines_test.tctA
}
