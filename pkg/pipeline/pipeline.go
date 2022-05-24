package pipeline

type Pipelineable[T any] interface {
	PipelineChan() chan T
	Close()
}
