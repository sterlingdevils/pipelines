package containerpipe_test

import (
	"fmt"
	"time"

	"github.com/sterlingdevils/pipelines/pkg/containerpipe"
)

type KeyType int
type DataType string

// Node does uses a pointer receiver on Key,  this is to test pointer things
type Node struct {
	key  KeyType
	data DataType
}

func (n *Node) Key() KeyType {
	return n.key
}

// Node2 does not use a pointer receiver on Key,  this is to test non-pointer things
type Node2 struct {
	key KeyType
}

func (n Node2) Key() KeyType {
	return n.key
}

// readAndPrint is used to get num items from the output channel and display them
func readAndPrint(num int, c <-chan *Node) {
	for i := 0; i < num; i++ {
		n := <-c
		fmt.Printf("%v, %v\n", n.key, n.data)
	}
}

func ExampleNew() {
	_, _ = containerpipe.New[KeyType, Node2]()
	// Output:
}

// ExampleContainerPipe_Close
func ExampleContainerPipe_Close() {
	r, _ := containerpipe.New[KeyType, Node2]()
	r.Close()
	// Output:
}

// ExampleContainerPipe_InChan
func ExampleContainerPipe_InChan() {
	r, _ := containerpipe.New[KeyType, *Node]()
	r.InChan() <- &Node{key: 7, data: "This is a test"}
	r.Close()
	// Output:
}

func ExampleContainerPipe_testtwoitemsinorder() {
	r, _ := containerpipe.New[KeyType, *Node]()
	r.InChan() <- &Node{key: 1, data: "I don't care what it is"}
	r.InChan() <- &Node{key: 2, data: "This is a test"}

	readAndPrint(2, r.OutChan())

	r.Close()
	// Output:
	// 1, I don't care what it is
	// 2, This is a test
}

func ExampleContainerPipe_testdeloffirst() {
	r, _ := containerpipe.New[KeyType, *Node]()
	r.InChan() <- &Node{key: 1, data: "I don't care what it is"}
	r.InChan() <- &Node{key: 2, data: "This is a test"}
	r.InChan() <- &Node{key: 3, data: "This is a test again"}
	r.DelChan() <- 1

	readAndPrint(2, r.OutChan())

	r.Close()
	// Output:
	// 2, This is a test
	// 3, This is a test again
}

func ExampleContainerPipe_testdelofsecond() {
	r, _ := containerpipe.New[KeyType, *Node]()
	r.InChan() <- &Node{key: 1, data: "I don't care what it is"}
	r.InChan() <- &Node{key: 2, data: "This is a test"}
	r.InChan() <- &Node{key: 3, data: "This is a test again"}
	r.DelChan() <- 2

	readAndPrint(2, r.OutChan())

	r.Close()
	// Output:
	// 1, I don't care what it is
	// 3, This is a test again
}

func ExampleContainerPipe_testdelonNotThere() {
	r, _ := containerpipe.New[KeyType, *Node]()
	r.InChan() <- &Node{key: 1, data: "I don't care what it is"}
	r.InChan() <- &Node{key: 2, data: "This is a test"}
	r.DelChan() <- 3

	readAndPrint(2, r.OutChan())

	r.Close()
	// Output:
	// 1, I don't care what it is
	// 2, This is a test
}

func ExampleContainerPipe_duptest() {
	r, _ := containerpipe.New[KeyType, *Node]()
	r.InChan() <- &Node{key: 1, data: "I don't care what it is"}
	// This should be dropped as a dup
	r.InChan() <- &Node{key: 1, data: "This is a test"}

	readAndPrint(1, r.OutChan())

	r.Close()
	// Output:
	// 1, I don't care what it is
}

// This example will test if we pass pointer fully thru the container
func ExampleContainerPipe_fullpointers() {
	// Notice that T is a pointer to a Node
	r, _ := containerpipe.New[KeyType, *Node]()

	ni := &Node{key: 1, data: "I don't care what it is"}

	r.InChan() <- ni
	no := <-r.OutChan()

	// Ok, we have our input node and the output after it went thru the container
	// lets change the key on the input node and make sure it changed on the output
	ni.key = 2
	fmt.Println(no.key)

	r.Close()
	// Output: 2
}

// // This example will test if we dont pass pointer fully thru the container
func ExampleContainerPipe_fullnopointers() {
	// Notice the small difference in T, we are no longer a pointer to Node
	r, _ := containerpipe.New[KeyType, Node2]()

	ni := Node2{key: 1}

	r.InChan() <- ni
	no := <-r.OutChan()

	// Ok, we have our input node and the output after it went thru the container
	// lets change the key on the input node and make sure the output is not changed
	ni.key = 2
	fmt.Println(no.key)

	r.Close()
	// Output: 1
}

// Checks the ApproxSize that it returns something close
func ExampleContainerPipe_ApproxSize() {
	numwrite := 100
	r, _ := containerpipe.New[KeyType, *Node]()

	s1 := r.ApproxSize()

	r.InChan() <- &Node{key: 1}
	s2 := r.ApproxSize()

	<-r.OutChan()
	time.Sleep(10 * time.Millisecond) // Have to wait a little for mainloop to cycle
	s3 := r.ApproxSize()

	for i := 0; i < numwrite; i++ {
		r.InChan() <- &Node{key: KeyType(i)}
	}
	time.Sleep(10 * time.Millisecond) // Have to wait a little for mainloop to cycle
	s4 := r.ApproxSize()

	fmt.Println(s1, s2, s3, s4)

	r.Close()
	// Output: 0 1 0 100
}