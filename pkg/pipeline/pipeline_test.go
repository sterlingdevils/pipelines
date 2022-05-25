package pipeline_test

import "github.com/sterlingdevils/pipelines/pkg/pipeline"

type Node[T any] struct {
}

func (n Node[T]) PipelineChan() chan T {
	return nil
}

func (n Node[T]) Close() {
}

// Check we can make a Pipelineable
func Example() {
	var p pipeline.Pipelineable[int] = Node[int]{}
	_ = p
	// Output:
}
