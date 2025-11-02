package pipeline

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// Stage represents a processing stage in a pipeline.
type Stage[I any, O any] struct {
	Name     string
	Workers  int
	Buffer   int
	Function func(Message[I]) (Message[O], error)
}

func (s *Stage[I, O]) Run(ctx context.Context, input <-chan Message[I]) (<-chan Message[O], *errgroup.Group) {
	output := make(chan Message[O], s.Buffer)
	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < s.Workers; i++ {
		eg.Go(func() error {
			for msg := range input {
				o, err := s.Function(msg)
				if err != nil {
					return fmt.Errorf("[%s]: %w", s.Name, err)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case output <- o:
				}
			}
			return nil
		})
	}

	go func() {
		_ = eg.Wait()
		close(output)
	}()

	return output, eg
}
