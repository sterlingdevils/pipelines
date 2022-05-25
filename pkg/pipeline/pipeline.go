package pipeline

type Pipelineable[T any] interface {
	// OutChan will return the output channel of the entire pipeline, This should be from the last component
	OutChan() <-chan T

	// PipelinChan needs to return a R/W chan that is for the output channel, This is for chaining the pipeline
	PipelineChan() chan T

	// Close is used to clear up any resources made by the component
	Close()
}