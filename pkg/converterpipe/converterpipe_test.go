package converterpipe_test

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/sterlingdevils/pipelines/pkg/converterpipe"
)

func Example() {
	cvt, err := converterpipe.New(
		func(i int) string {
			return strconv.Itoa(i)
		})

	if err != nil {
		log.Fatalln(err)
	}

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o, reflect.TypeOf(o))

	// Output:
	// 5 string
}

// ignore the input and just return hello
func returnHello(i int) string {
	return "Hello"
}

func Example_fixedoutput() {
	cvt, _ := converterpipe.New(returnHello)

	cvt.InChan() <- 5
	o := <-cvt.OutChan()

	fmt.Println(o)

	// Output:
	// Hello
}
