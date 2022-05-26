package converterpipe_test

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/sterlingdevils/pipelines/pkg/converterpipe"
)

func Example() {
	cvt := converterpipe.New(
		func(i int) string {
			return strconv.Itoa(i)
		})

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o, reflect.TypeOf(o))

	// Output:
	// 5 string
}

func Example_numchar() {
	cvt := converterpipe.New(
		func(i int) string {
			return fmt.Sprintf("%010d", i)
		})

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o)

	// Output:
	// 0000000005
}

// ignore the input and just return hello
func returnHello(i int) string {
	return "Hello"
}

func Example_fixedoutput() {
	cvt := converterpipe.New(returnHello)

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o)

	// Output:
	// Hello
}
