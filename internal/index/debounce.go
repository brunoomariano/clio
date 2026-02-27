package index

import (
	"context"
	"time"
)

type DebouncedExecutor[T any] struct {
	Delay time.Duration
}

type DebouncedResult[T any] struct {
	Value T
	Err   error
}

func (d DebouncedExecutor[T]) Run(ctx context.Context, fn func(context.Context) (T, error)) <-chan DebouncedResult[T] {
	ch := make(chan DebouncedResult[T], 1)
	go func() {
		defer close(ch)
		select {
		case <-time.After(d.Delay):
			if ctx.Err() != nil {
				return
			}
			val, err := fn(ctx)
			if ctx.Err() != nil {
				return
			}
			ch <- DebouncedResult[T]{Value: val, Err: err}
		case <-ctx.Done():
			return
		}
	}()
	return ch
}
