package main

import (
	"context"
	"fmt"

	"github.com/aawadall/go-concurrency-patterns/cmd/pipelines/pipeline"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Source channel
	input := make(chan pipeline.Message[int])

	go func() {
		for i := 1; i <= 10; i++ {
			input <- pipeline.Message[int]{ID: int64(i), Payload: i}
		}
		close(input)
	}()

	// define stages
	squareStage := pipeline.Stage[int, int]{
		Name:    "Square Stage",
		Workers: 3,
		Buffer:  4,
		Function: func(x pipeline.Message[int]) (pipeline.Message[int], error) {
			return pipeline.Message[int]{ID: x.ID, Payload: x.Payload * x.Payload}, nil
		},
	}

	doubleStage := pipeline.Stage[int, int]{
		Name:    "Double Stage",
		Workers: 2,
		Buffer:  4,
		Function: func(x pipeline.Message[int]) (pipeline.Message[int], error) {
			return pipeline.Message[int]{ID: x.ID, Payload: x.Payload * 2}, nil
		},
	}

	// wire up the pipeline
	out1, g1 := squareStage.Run(ctx, input)
	out2, g2 := doubleStage.Run(ctx, out1)

	// drain the output
	go func() {
		_ = g1.Wait()
		_ = g2.Wait()
	}()

	for result := range out2 {
		println(fmt.Sprintf("[%d]: %d", result.ID, result.Payload))
	}
}
