package pipelines_test

import "github.com/sterlingdevils/pipelines"

type Node[T any] struct {
}

func (n Node[T]) PipelineChan() chan T {
	return nil
}

func (n Node[T]) Close() {
}

// Check we can make a Pipelineable
func Example() {
	var p pipelines.Pipeliner[int] = Node[int]{}
	_ = p
	// Output:
}
