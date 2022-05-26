package retrypipe_test

import (
	"context"
	"fmt"
	"time"

	"github.com/sterlingdevils/gobase/pkg/serialnum"
	"github.com/sterlingdevils/pipelines/retrypipe"
)

type KeyType uint64
type DataType []byte
type Obj struct {
	Sn  KeyType
	ctx context.Context
	can context.CancelFunc

	Data DataType
}

// Context returns the private context
func (o Obj) Context() context.Context {
	return o.ctx
}

func (o *Obj) Key() KeyType {
	return o.Sn
}

func NewObj(timeout time.Duration) (*Obj, error) {
	c, cancel := context.WithTimeout(context.Background(), timeout)
	o := Obj{ctx: c, can: cancel}

	return &o, nil
}

func Example() {
	retry := retrypipe.New[KeyType, *Obj]()

	retry.Close()
	// Output:
}

func ExampleRetry_inout() {
	sn := serialnum.New()
	retry := retrypipe.New[KeyType, *Obj]()

	go func() {
		for o := range retry.OutChan() {
			fmt.Println(o.Key())
		}
	}()

	for i := 0; i < 10; i++ {
		o, _ := NewObj(2 * time.Second)
		o.Sn = KeyType(sn.Next())
		retry.InChan() <- o
	}

	time.Sleep(3 * time.Second)
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
}

func ExampleRetry_pointercheck() {
	retry := retrypipe.New[KeyType, *Obj]()

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

func ExampleRetry_Close() {
	retry := retrypipe.New[KeyType, *Obj]()
	retry.Close()
	// Output:
}
