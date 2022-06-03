package pipelines_test

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/sterlingdevils/pipelines"
)

func ExampleConverterPipe_New() {
	cvt := pipelines.ConverterPipe[int, string]{}.New(
		func(i int) (string, error) {
			return strconv.Itoa(i), nil
		})

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o, reflect.TypeOf(o))

	// Output:
	// 5 string
}

func ExampleConverterPipe_numchar() {
	cvt := pipelines.ConverterPipe[int, string]{}.New(
		func(i int) (string, error) {
			return fmt.Sprintf("%010d", i), nil
		})

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o)

	// Output:
	// 0000000005
}

// ignore the input and just return hello
func returnHello(i int) (string, error) {
	return "Hello", nil
}

func ExampleConverterPipe_fixedoutput() {
	cvt := pipelines.ConverterPipe[int, string]{}.New(returnHello)

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o)

	// Output:
	// Hello
}
