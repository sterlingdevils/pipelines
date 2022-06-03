package typeconverterpipe_test

import (
	"fmt"
	"reflect"

	"github.com/sterlingdevils/pipelines/typeconverterpipe"
)

type Iface interface {
	Boo() int
}

type A struct{}

func (a A) Boo() int {
	return 2
}

type B struct{}

func (b B) Boo() int {
	return 3
}

func Example() {
	cvt := typeconverterpipe.New[A, Iface]()

	cvt.InChan() <- A{}
	o := <-cvt.OutChan()

	fmt.Println(o.Boo(), reflect.TypeOf(o))

	// Output:
	// 5 int
}
