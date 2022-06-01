package pipelines

type PipelineChaner[T any] interface {
	// PipelineChan needs to return a R/W chan that is for the output channel, This is for chaining the pipeline
	PipelineChan() chan T
}

type Closer interface {
	// Close is used to clear up any resources made by the component
	Close()
}

type Pipeline[T any] interface {
	PipelineChaner[T]
	Closer
}

type Dataer interface {
	// Data returns a byte slice to the data it holds
	Data() []byte
}

type Keyer[K comparable] interface {
	Key() K
}

type FileNamer interface {
	FileName() string
}

type FileNamerDataer interface {
	Dataer
	FileNamer
}
