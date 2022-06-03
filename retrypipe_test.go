package pipelines_test

import (
	"fmt"
	"time"

	"github.com/sterlingdevils/gobase/serialnum"
	"github.com/sterlingdevils/pipelines"
)

type rptKeyType uint64
type rptDataType []byte

type Obj struct {
	Sn   rptKeyType
	Data rptDataType
}

func (o *Obj) Key() rptKeyType {
	return o.Sn
}

func NewObj(timeout time.Duration) (*Obj, error) {
	return &Obj{}, nil
}

func ExampleRetryPipe() {
	retry := pipelines.RetryPipe[rptKeyType, *Obj]{}.New()

	retry.Close()
	// Output:
}

func ExampleRetryPipe_inout() {
	sn := serialnum.New()
	retry := pipelines.RetryPipe[rptKeyType, *Obj]{}.New()

	go func() {
		for o := range retry.OutChan() {
			fmt.Println(o.Key())
		}
	}()

	for i := 0; i < 10; i++ {
		o, _ := NewObj(2 * time.Second)
		o.Sn = rptKeyType(sn.Next())
		retry.InChan() <- o
	}

	time.Sleep(4 * time.Second)
	retry.Close()

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
}

func ExampleRetryPipe_pointercheck() {
	retry := pipelines.RetryPipe[rptKeyType, *Obj]{}.New()

	// Check that we are passing pointer
	o, _ := NewObj(5 * time.Second)
	retry.InChan() <- o
	o.Sn = 5
	go func() {
		for o := range retry.OutChan() {
			fmt.Println(o.Key())
		}
	}()

	retry.InChan() <- o

	time.Sleep(2 * time.Second)
	retry.Close()

	// Output:
	// 5
	// 5
}

func ExampleRetryPipe_Close() {
	retry := pipelines.RetryPipe[rptKeyType, *Obj]{}.New()
	retry.Close()
	// Output:
}
