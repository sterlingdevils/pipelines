package pipeline_test

import (
	"log"

	"github.com/sterlingdevils/pipelines/pkg/bufferpipe"
)

type Node struct {
	Name string
}

func Example_betterpipeline() {
	b1, err := bufferpipe.New[any](10)
	if err != nil {
		log.Fatal(err)
	}

	b2, err := bufferpipe.NewWithPipeline[any](5, b1)
	if err != nil {
		log.Fatal(err)
	}

	b2.Close()
	// Output:
}
