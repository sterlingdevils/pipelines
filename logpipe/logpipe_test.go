/*
  NOTE:  White Box testing to have access to inchan and outchan
*/
package logpipe

import "fmt"

func Example() {
	lg := New[int]("example")
	lg.inchan <- 5
	fmt.Println(<-lg.outchan)
	// Output:
	// 5
}
