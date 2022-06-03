package pipelines

const (
	CHANSIZE = 0
)

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

type Sizer interface {
	Size() int
}

type Dataer interface {
	// Data returns a byte slice to the data it holds
	Data() []byte
}

type DataSizer interface {
	Dataer
	Sizer
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

// RecoverFromClosedChan is used when it is OK if the channel is closed we are writing on
// This is not great using the string compare but the go runtime uses a generic error so we
// can't trap this any other way.
func recoverFromClosedChan() {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok && e.Error() == "send on closed channel" {
			return
		}
		panic(r)
	}
}
