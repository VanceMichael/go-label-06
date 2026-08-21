package stream

import (
	"context"
	"sync"
)

type subscriber[T any] struct {
	inbox  chan T
	output chan T
	done   chan struct{}
	once   sync.Once
}

func newSubscriber[T any](buffer int) *subscriber[T] {
	state := &subscriber[T]{
		inbox:  make(chan T, buffer),
		output: make(chan T, buffer),
		done:   make(chan struct{}),
	}
	go state.relay()
	return state
}

func (state *subscriber[T]) relay() {
	defer close(state.output)
	for {
		select {
		case <-state.done:
			return
		case value := <-state.inbox:
			if !state.forward(value) {
				return
			}
		}
	}
}

func (state *subscriber[T]) forward(value T) bool {
	select {
	case state.output <- value:
		return true
	case <-state.done:
		return false
	}
}

func (state *subscriber[T]) enqueue(ctx context.Context, value T) error {
	select {
	case <-state.done:
		return nil
	default:
	}

	select {
	case state.inbox <- value:
		return nil
	case <-state.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (state *subscriber[T]) close() {
	state.once.Do(func() {
		close(state.done)
	})
}
