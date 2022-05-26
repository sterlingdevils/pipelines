package pipeliner_test

import "github.com/sterlingdevils/pipelines/pipeliner"

type Node[T any] struct {
}

func (n Node[T]) PipelineChan() chan T {
	return nil
}

func (n Node[T]) Close() {
}

// Check we can make a Pipelineable
func Example() {
	var p pipeliner.Pipeliner[int] = Node[int]{}
	_ = p
	// Output:
}
