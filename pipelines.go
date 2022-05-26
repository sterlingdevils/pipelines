package pipelines

type Pipeliner[T any] interface {
	// PipelineChan needs to return a R/W chan that is for the output channel, This is for chaining the pipeline
	PipelineChan() chan T

	// Close is used to clear up any resources made by the component
	Close()
}
